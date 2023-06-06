package net

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/decoder"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/baix/pack"
	"github.com/oldbai555/lbtool/log"
	"github.com/oldbai555/lbtool/pkg/routine"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"time"
)

var _ iface.IServer = (*Server)(nil)

type Server struct {
	Name string

	IPVersion string

	IP string

	Port int

	WsPort int

	msgHandler iface.IMsgHandle

	ConnMgr iface.IConnManager

	onConnStart func(conn iface.IConnection) // conn HOOK 前置钩子

	onConnStop func(conn iface.IConnection) // conn HOOK 后置钩子

	packet iface.IDataPack // 数据报文封包

	exitChan chan struct{} // 异步捕捉链接关闭状态

	decoder iface.IDecoder // 断粘包解码器

	hc iface.IHeartbeatChecker // 心跳检测器

	upgrader *websocket.Upgrader // websocket

	websocketAuth func(r *http.Request) error // websocket connection authentication

	cID uint64
}

func newServerWithConfig(config *conf.Config, ipVersion string) iface.IServer {
	s := &Server{
		Name:       config.Name,
		IPVersion:  ipVersion,
		IP:         config.Host,
		Port:       config.TCPPort,
		WsPort:     config.WsPort,
		msgHandler: newMsgHandle(),
		ConnMgr:    newConnManager(),
		packet:     pack.Factory().NewPack(iface.BaiXDataPack),
		decoder:    decoder.NewTLVDecoder(),
		upgrader: &websocket.Upgrader{
			ReadBufferSize: int(config.IOReadBuffSize),
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}

	return s
}

func NewServer() iface.IServer {
	conf.GlobalConfig.Show()
	conf.GlobalConfig.Mode = conf.ServerModeTcp
	return newServerWithConfig(conf.GlobalConfig, conf.ServerModeTcp)
}

func NewWsServer() iface.IServer {
	conf.GlobalConfig.Show()
	conf.GlobalConfig.Mode = conf.ServerModeWebsocket
	return newServerWithConfig(conf.GlobalConfig, conf.ServerModeWebsocket)
}

func (s *Server) Start() {
	log.Infof("【BaiX】[START] Server name: %s,listener at IP: %s, Port %d is starting", s.Name, s.IP, s.Port)
	s.exitChan = make(chan struct{})

	// 将解码器添加到拦截器
	if s.decoder != nil {
		s.msgHandler.AddInterceptor(s.decoder)
	}

	// 启动worker工作池机制
	s.msgHandler.StartWorkerPool()

	// 开启一个 go 去做服务端 Listener 业务
	switch conf.GlobalConfig.Mode {
	case conf.ServerModeTcp:
		routine.Go(context.TODO(), func(ctx context.Context) error {
			err := s.ListenTcpConn()
			if err != nil {
				log.Errorf("err is %v", err)
				return err
			}
			return nil
		})
	case conf.ServerModeWebsocket:
		routine.Go(context.TODO(), func(ctx context.Context) error {
			err := s.ListenWebsocketConn()
			if err != nil {
				log.Errorf("err is %v", err)
				return err
			}
			return nil
		})
	}
}

func (s *Server) ListenTcpConn() error {
	// 1. Get a TCP address
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		log.Errorf("【BaiX】[START] resolve tcp addr err: %v", err)
		return err
	}

	// 2. Listen to the server address
	var listener net.Listener
	switch {
	case conf.GlobalConfig.CertFile != "" && conf.GlobalConfig.PrivateKeyFile != "":
		// Read certificate and private key
		crt, err := tls.LoadX509KeyPair(conf.GlobalConfig.CertFile, conf.GlobalConfig.PrivateKeyFile)
		if err != nil {
			log.Errorf("err is %v", err)
			return err
		}

		// TLS connection
		tlsConfig := &tls.Config{}
		tlsConfig.Certificates = []tls.Certificate{crt}
		tlsConfig.Time = time.Now
		tlsConfig.Rand = rand.Reader
		listener, err = tls.Listen(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port), tlsConfig)
		if err != nil {
			log.Errorf("err is %v", err)
			return err
		}
	default:
		listener, err = net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			log.Errorf("err is %v", err)
			return err
		}
	}

	// 3. Start server network connection business
	routine.Go(context.TODO(), func(ctx context.Context) error {
		for {
			// 3.1 Set the maximum connection control for the server. If it exceeds the maximum connection, wait.
			// (设置服务器最大连接控制,如果超过最大连接，则等待)
			if s.ConnMgr.Len() >= conf.GlobalConfig.MaxConn {
				log.Errorf("【BaiX】Exceeded the maxConnNum:%d, Wait:%d", conf.GlobalConfig.MaxConn, AcceptDelay.duration)
				AcceptDelay.Delay()
				continue
			}

			// 3.2 Block and wait for a client to establish a connection request.
			// (阻塞等待客户端建立连接请求)
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					log.Errorf("【BaiX】Listener closed")
					return err
				}
				log.Errorf("【BaiX】Accept err: %v", err)
				AcceptDelay.Delay()
				continue
			}

			AcceptDelay.Reset()

			// 3.4 Handle the business method for this new connection request. At this time, the handler and conn should be bound.
			// (处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的)
			newCid := atomic.AddUint64(&s.cID, 1)
			dealConn := newServerConn(s, conn, newCid)

			routine.Go(ctx, func(ctx context.Context) error {
				s.StartConn(dealConn)
				return nil
			})

		}
	})
	select {
	case <-s.exitChan:
		err := listener.Close()
		if err != nil {
			log.Errorf("err is %v", err)
			return err
		}
	}
	return nil
}

func (s *Server) ListenWebsocketConn() error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Check if the server has reached the maximum allowed number of connections
		// (设置服务器最大连接控制,如果超过最大连接，则等待)
		if s.ConnMgr.Len() >= conf.GlobalConfig.MaxConn {
			log.Errorf("【BaiX】Exceeded the maxConnNum:%d, Wait:%d", conf.GlobalConfig.MaxConn, AcceptDelay.duration)
			AcceptDelay.Delay()
			return
		}

		// 2. 如果需要 websocket 认证请设置认证信息
		if s.websocketAuth != nil {
			err := s.websocketAuth(r)
			if err != nil {
				log.Errorf("【BaiX】websocket auth err:%v", err)
				w.WriteHeader(401)
				AcceptDelay.Delay()
				return
			}
		}

		// 3. 判断 header 里面是有子协议
		if len(r.Header.Get("Sec-Websocket-Protocol")) > 0 {
			s.upgrader.Subprotocols = websocket.Subprotocols(r)
		}

		// 4. 升级成 websocket 连接
		conn, err := s.upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("new websocket err:%v", err)
			w.WriteHeader(500)
			AcceptDelay.Delay()
			return
		}
		AcceptDelay.Reset()

		// 5. 处理该新连接请求的 业务 方法， 此时应该有 handler 和 conn是绑定的
		newCid := atomic.AddUint64(&s.cID, 1)
		wsConn := newWebsocketConn(s, conn, newCid)
		routine.Go(context.TODO(), func(ctx context.Context) error {
			s.StartConn(wsConn)
			return nil
		})

	})

	err := http.ListenAndServe(fmt.Sprintf("%s:%d", s.IP, s.WsPort), nil)
	if err != nil {
		log.Errorf("【BaiX】err is %v", err)
		panic(err)
	}
	return nil
}

func (s *Server) StartConn(conn iface.IConnection) {
	// HeartBeat check
	if s.hc != nil {
		// Clone a heart-beat checker from the server side
		heartBeatChecker := s.hc.Clone()

		// Bind current connection
		heartBeatChecker.BindConn(conn)
	}

	// Start processing business for the current connection
	conn.Start()
}

func (s *Server) Stop() {
	// 将其他需要清理的连接信息或者其他信息 也要一并停止或者清理
	s.ConnMgr.ClearConn()
	s.exitChan <- struct{}{}
	close(s.exitChan)
}

func (s *Server) Serve() {
	s.Start()

	// 阻塞,否则主Go退出， listener 的go将会退出
	c := make(chan os.Signal, 1)

	// 监听指定信号 ctrl+c kill信号
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("【BaiX】[SERVE] BaiX server , name %s, Serve Interrupt, signal = %v", s.Name, sig)
}

func (s *Server) AddRouter(msgID uint32, router iface.IRouter) {
	s.msgHandler.AddRouter(msgID, router)
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

func (s *Server) SetOnConnStart(f func(iface.IConnection)) {
	s.onConnStart = f
}

func (s *Server) SetOnConnStop(f func(iface.IConnection)) {
	s.onConnStop = f
}

func (s *Server) GetOnConnStart() func(iface.IConnection) {
	return s.onConnStart
}

func (s *Server) GetOnConnStop() func(iface.IConnection) {
	return s.onConnStop
}

func (s *Server) GetPacket() iface.IDataPack {
	return s.packet
}

func (s *Server) GetMsgHandler() iface.IMsgHandle {
	return s.msgHandler
}

func (s *Server) SetPacket(packet iface.IDataPack) {
	s.packet = packet
}

func (s *Server) StartHeartBeat(interval time.Duration) {
	checker := NewHeartbeatChecker(interval)

	// Add the heartbeat check router. (添加心跳检测的路由)
	s.AddRouter(checker.MsgID(), checker.Router())

	// Bind the heartbeat checker to the server. (server绑定心跳检测器)
	s.hc = checker
}

func (s *Server) GetHeartBeat() iface.IHeartbeatChecker {
	return s.hc
}

func (s *Server) GetLengthField() *iface.LengthField {
	if s.decoder != nil {
		return s.decoder.GetLengthField()
	}
	return nil
}

func (s *Server) SetDecoder(iDecoder iface.IDecoder) {
	s.decoder = iDecoder
}

func (s *Server) AddInterceptor(interceptor iface.IInterceptor) {
	s.msgHandler.AddInterceptor(interceptor)
}

func (s *Server) SetWebsocketAuth(f func(r *http.Request) error) {
	s.websocketAuth = f
}

func (s *Server) ServerName() string {
	return s.Name
}
