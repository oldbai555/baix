package net

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/decoder"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/baix/pack"
	"github.com/oldbai555/lbtool/log"
	"github.com/oldbai555/lbtool/pkg/routine"
	"net"
	"time"
)

var _ iface.IClient = (*Client)(nil)

type Client struct {
	// Client Name 客户端的名称
	Name string
	// IP of the target server to connect 目标链接服务器的IP
	Ip string
	// Port of the target server to connect 目标链接服务器的端口
	Port int
	// Client version tcp,websocket,客户端版本 tcp,websocket
	version string
	// Connection instance 链接实例
	conn iface.IConnection
	// Hook function called on connection start 该client的连接创建时Hook函数
	onConnStart func(conn iface.IConnection)
	// Hook function called on connection stop 该client的连接断开时的Hook函数
	onConnStop func(conn iface.IConnection)
	// Data packet packer 数据报文封包方式
	packet iface.IDataPack
	// Asynchronous channel for capturing connection close status 异步捕获链接关闭状态
	exitChan chan struct{}
	// Message management module 消息管理模块
	msgHandler iface.IMsgHandle
	// Disassembly and assembly decoder for resolving sticky and broken packages
	//断粘包解码器
	decoder iface.IDecoder
	// Heartbeat checker 心跳检测器
	hc iface.IHeartbeatChecker
	// Use TLS 使用TLS
	useTLS bool
	// For websocket connections
	dialer *websocket.Dialer
	// Error channel
	ErrChan chan error
}

func NewClient(ip string, port int) iface.IClient {

	c := &Client{
		// Default name, can be modified using the WithNameClient Option
		// (默认名称，可以使用WithNameClient的Option修改)
		Name: "【BaiX】ClientTcp",
		Ip:   ip,
		Port: port,

		msgHandler: newMsgHandle(),
		packet:     pack.Factory().NewPack(iface.BaiXDataPack), // 默认使用 BaiX 的TLV封包方式
		decoder:    decoder.NewTLVDecoder(),                    // 默认使用 BaiX 的TLV解码器
		version:    conf.ServerModeTcp,
		ErrChan:    make(chan error),
	}

	return c
}

func NewWsClient(ip string, port int) iface.IClient {

	c := &Client{
		// Default name, can be modified using the WithNameClient Option
		// (默认名称，可以使用WithNameClient的Option修改)
		Name: "【BaiX】ClientWs",
		Ip:   ip,
		Port: port,

		msgHandler: newMsgHandle(),
		packet:     pack.Factory().NewPack(iface.BaiXDataPack), // 默认使用 BaiX 的TLV封包方式
		decoder:    decoder.NewTLVDecoder(),                    // 默认使用 BaiX 的TLV解码器
		version:    conf.ServerModeWebsocket,
		dialer:     &websocket.Dialer{},
		ErrChan:    make(chan error),
	}

	return c
}

func NewTLSClient(ip string, port int) iface.IClient {

	c, _ := NewClient(ip, port).(*Client)

	c.useTLS = true

	return c
}

func (c *Client) Restart() {
	c.exitChan = make(chan struct{})

	// 客户端将协程池关闭
	conf.GlobalConfig.WorkerPoolSize = 0

	routine.Go(context.Background(), func(ctx context.Context) error {
		addr := &net.TCPAddr{
			IP:   net.ParseIP(c.Ip),
			Port: c.Port,
			Zone: "", //for ipv6, ignore
		}

		// 创建原始Socket，得到net.Conn
		switch c.version {
		case "websocket":
			wsAddr := fmt.Sprintf("ws://%s:%d", c.Ip, c.Port)

			// 创建原始Socket，得到net.Conn
			wsConn, _, err := c.dialer.Dial(wsAddr, nil)
			if err != nil {
				// Conection failed
				log.Errorf("【BaiX】WsClient connect to server failed, err:%v", err)
				c.ErrChan <- err
				return err
			}
			// Create Connection object
			c.conn = newWsClientConn(c, wsConn)
		case "tcp":
			var conn net.Conn
			var err error
			if c.useTLS {
				// TLS encryption
				config := &tls.Config{
					// 这里是跳过证书验证，因为证书签发机构的CA证书是不被认证的
					InsecureSkipVerify: true,
				}

				conn, err = tls.Dial("tcp", fmt.Sprintf("%v:%v", net.ParseIP(c.Ip), c.Port), config)
				if err != nil {
					log.Errorf("【BaiX】tls client connect to server failed, err:%v", err)
					c.ErrChan <- err
					return err
				}
			} else {
				conn, err = net.DialTCP("tcp", nil, addr)
				if err != nil {
					// Conection failed
					log.Errorf("【BaiX】client connect to server failed, err:%v", err)
					c.ErrChan <- err
					return err
				}
			}
			// Create Connection object
			c.conn = newClientConn(c, conn)
		}

		log.Infof("【BaiX】[START] Client LocalAddr: %s, RemoteAddr: %s\n", c.conn.LocalAddr(), c.conn.RemoteAddr())

		// HeartBeat detection
		if c.hc != nil {
			// 创建链接成功，绑定链接与心跳检测器
			c.hc.BindConn(c.conn)
		}

		// Start connection
		routine.Go(context.Background(), func(ctx context.Context) error {
			c.conn.Start()
			return nil
		})

		select {
		case <-c.exitChan:
			log.Infof("【BaiX】client exit.")
		}
		return nil
	})
}

func (c *Client) Start() {
	// 将解码器添加到拦截器
	if c.decoder != nil {
		c.msgHandler.AddInterceptor(c.decoder)
	}

	c.Restart()
}

func (c *Client) Stop() {
	log.Infof("【BaiX】[STOP] BaiX Client LocalAddr: %s, RemoteAddr: %s", c.conn.LocalAddr(), c.conn.RemoteAddr())
	c.conn.Stop()
	c.exitChan <- struct{}{}
	close(c.exitChan)
	close(c.ErrChan)
}

func (c *Client) AddRouter(msgID uint32, router iface.IRouter) {
	if c.msgHandler != nil {
		c.msgHandler.AddRouter(msgID, router)
	}
}

func (c *Client) Conn() iface.IConnection {
	return c.conn
}

func (c *Client) SetOnConnStart(f func(iface.IConnection)) {
	c.onConnStart = f
}

func (c *Client) SetOnConnStop(f func(iface.IConnection)) {
	c.onConnStop = f
}

func (c *Client) GetOnConnStart() func(iface.IConnection) {
	return c.onConnStart
}

func (c *Client) GetOnConnStop() func(iface.IConnection) {
	return c.onConnStop
}

func (c *Client) GetPacket() iface.IDataPack {
	return c.packet
}

func (c *Client) SetPacket(pack iface.IDataPack) {
	c.packet = pack
}

func (c *Client) GetMsgHandler() iface.IMsgHandle {
	return c.msgHandler
}

func (c *Client) StartHeartBeat(duration time.Duration) {
	checker := NewHeartbeatChecker(duration)
	c.AddRouter(checker.MsgID(), checker.Router())
	c.hc = checker
}

func (c *Client) GetLengthField() *iface.LengthField {
	if c.decoder != nil {
		return c.decoder.GetLengthField()
	}
	return nil
}

func (c *Client) SetDecoder(decoder iface.IDecoder) {
	c.decoder = decoder
}

func (c *Client) AddInterceptor(interceptor iface.IInterceptor) {
	if c.msgHandler != nil {
		c.msgHandler.AddInterceptor(interceptor)
	}
}

func (c *Client) GetErrChan() chan error {
	return c.ErrChan
}

func (c *Client) SetName(s string) {
	c.Name = s
}

func (c *Client) GetName() string {
	return c.Name
}
