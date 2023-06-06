package net

import (
	"context"
	"encoding/hex"
	"github.com/gorilla/websocket"
	"github.com/oldbai555/baix/berr"
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/baix/interceptor"
	"github.com/oldbai555/baix/pack"
	"github.com/oldbai555/lbtool/log"

	"github.com/oldbai555/lbtool/pkg/routine"
	"net"
	"sync"
	"time"
)

var _ iface.IConnection = (*WsConnection)(nil)

type WsConnection struct {
	// conn 当前连接的socket TCP套接字
	conn *websocket.Conn

	// connID 当前连接的ID 也可以称作为SessionID，ID全局唯一 ，服务端Connection使用
	// uint64 取值范围：0 ~ 18,446,744,073,709,551,615
	// 这个是理论支持的进程connID的最大数量
	connID uint64

	// msgHandler 消息管理MsgID和对应处理方法的消息管理模块
	msgHandler iface.IMsgHandle

	// ctx and cancel 告知该链接已经退出/停止的channel
	ctx    context.Context
	cancel context.CancelFunc

	// msgBuffChan 有缓冲管道，用于读、写两个goroutine之间的消息通信
	msgBuffChan chan []byte

	// msgLock 用户收发消息的Lock
	msgLock sync.RWMutex

	// property 链接属性
	property map[string]interface{}

	// propertyLock 保护当前property的锁
	propertyLock sync.Mutex

	// isClosed 当前连接的关闭状态
	isClosed bool

	// connManager 当前链接是属于哪个Connection Manager的
	connManager iface.IConnManager

	// onConnStart 当前连接创建时Hook函数
	onConnStart func(conn iface.IConnection)

	// onConnStop 当前连接断开时的Hook函数
	onConnStop func(conn iface.IConnection)

	// packet 数据报文封包方式
	packet iface.IDataPack

	// lastActivityTime 最后一次活动时间
	lastActivityTime time.Time

	// frameDecoder 断粘包解码器
	frameDecoder iface.IFrameDecoder

	// hc 心跳检测器
	hc iface.IHeartbeatChecker

	// name 链接名称，默认与创建链接的Server/Client的Name一致
	name string

	// localAddr 当前链接的本地地址
	localAddr string

	// remoteAddr 当前链接的远程地址
	remoteAddr string
}

func newWebsocketConn(server iface.IServer, conn *websocket.Conn, connID uint64) iface.IConnection {
	// Initialize Conn properties (初始化Conn属性)
	c := &WsConnection{
		conn:        conn,
		connID:      connID,
		isClosed:    false,
		msgBuffChan: nil,
		property:    nil,
		name:        server.ServerName(),
		localAddr:   conn.LocalAddr().String(),
		remoteAddr:  conn.RemoteAddr().String(),
	}

	lengthField := server.GetLengthField()
	if lengthField != nil {
		c.frameDecoder = interceptor.NewFrameDecoder(*lengthField)
	}

	// 从server继承过来的属性
	c.packet = server.GetPacket()
	c.onConnStart = server.GetOnConnStart()
	c.onConnStop = server.GetOnConnStop()
	c.msgHandler = server.GetMsgHandler()

	// 将当前的Connection与Server的ConnManager绑定
	c.connManager = server.GetConnMgr()

	// 将新创建的Conn添加到链接管理中
	server.GetConnMgr().Add(c)

	return c
}

func newWsClientConn(client iface.IClient, conn *websocket.Conn) iface.IConnection {
	c := &WsConnection{
		conn:        conn,
		connID:      0, //client ignore
		isClosed:    false,
		msgBuffChan: nil,
		property:    nil,
		name:        client.GetName(),
		localAddr:   conn.LocalAddr().String(),
		remoteAddr:  conn.RemoteAddr().String(),
	}

	lengthField := client.GetLengthField()
	if lengthField != nil {
		c.frameDecoder = interceptor.NewFrameDecoder(*lengthField)
	}

	// Inherit properties from client (从client继承过来的属性)
	c.packet = client.GetPacket()
	c.onConnStart = client.GetOnConnStart()
	c.onConnStop = client.GetOnConnStop()
	c.msgHandler = client.GetMsgHandler()

	return c
}

// StartWriter 写消息Goroutine， 用户将数据发送给客户端
func (w *WsConnection) StartWriter() {
	log.Infof("【BaiX】Writer Goroutine is running")
	defer func() {
		log.Infof("【BaiX】%s [conn Writer exit!]", w.RemoteAddr().String())
	}()

	for {
		select {
		case data, ok := <-w.msgBuffChan:
			if ok {
				if err := w.conn.WriteMessage(websocket.BinaryMessage, data); err != nil {
					log.Errorf("【BaiX】Send Buff Data error:, %s Conn Writer exit", err)
					break
				}
			} else {
				log.Errorf("【BaiX】msgBuffChan is Closed")
				break
			}
		case <-w.ctx.Done():
			return
		}
	}
}

// StartReader 读消息Goroutine，用于从客户端中读取数据
func (w *WsConnection) StartReader() {
	log.Infof("【BaiX】[Reader Goroutine is running]")
	defer func() {
		log.Infof("【BaiX】%s [conn Reader exit!]", w.RemoteAddr().String())
		w.Stop()
	}()

	for {
		select {
		case <-w.ctx.Done():
			return
		default:
			// 从conn的IO中读取数据到内存缓冲buffer中
			messageType, buffer, err := w.conn.ReadMessage()
			if err != nil {
				w.cancel()
				return
			}

			if messageType == websocket.PingMessage {
				w.updateActivity()
				continue
			}

			n := len(buffer)
			log.Debugf("【BaiX】read buffer %s ", hex.EncodeToString(buffer[0:n]))

			// 正常读取到对端数据，更新心跳检测Active状态
			if n > 0 && w.hc != nil {
				w.updateActivity()
			}

			// 处理自定义协议断粘包问题
			if w.frameDecoder != nil {
				// 为读取到的0-n个字节的数据进行解码
				bufArrays := w.frameDecoder.Decode(buffer)
				if bufArrays == nil {
					continue
				}
				for _, bytes := range bufArrays {
					log.Debugf("【BaiX】read buffer %s ", hex.EncodeToString(bytes))
					msg := pack.NewMessage(uint32(len(bytes)), bytes)

					// 得到当前客户端请求的Request数据
					req := NewRequest(w, msg)
					w.msgHandler.Execute(req)
				}
				continue
			}

			msg := pack.NewMessage(uint32(n), buffer[0:n])
			// 得到当前客户端请求的Request数据
			req := NewRequest(w, msg)
			w.msgHandler.Execute(req)

		}
	}
}

func (w *WsConnection) updateActivity() {
	w.lastActivityTime = time.Now()
}

func (w *WsConnection) callOnConnStart() {
	if w.onConnStart != nil {
		log.Infof("【BaiX】 CallOnConnStart....")
		w.onConnStart(w)
	}
}

func (w *WsConnection) callOnConnStop() {
	if w.onConnStop != nil {
		log.Infof("【BaiX】 CallOnConnStop....")
		w.onConnStop(w)
	}
}

func (w *WsConnection) finalizer() {
	// Call the callback function registered by the user when closing the connection if it exists
	// (如果用户注册了该链接的	关闭回调业务，那么在此刻应该显示调用)
	w.callOnConnStop()

	w.msgLock.Lock()
	defer w.msgLock.Unlock()

	// If the connection has already been closed
	if w.isClosed == true {
		return
	}

	// Stop the heartbeat detector associated with the connection
	if w.hc != nil {
		w.hc.Stop()
	}

	// Close the socket connection
	_ = w.conn.Close()

	// Remove the connection from the connection manager
	if w.connManager != nil {
		w.connManager.Remove(w)
	}

	// Close all channels associated with the connection
	if w.msgBuffChan != nil {
		close(w.msgBuffChan)
	}

	w.isClosed = true

	log.Infof("【BaiX】Conn Stop()...ConnID = %d", w.connID)
}

func (w *WsConnection) Start() {
	w.ctx, w.cancel = context.WithCancel(context.Background())
	// 按照用户传递进来的创建连接时需要处理的业务，执行钩子方法
	w.callOnConnStart()

	// 启动心跳检测
	if w.hc != nil {
		w.hc.Start()
		w.updateActivity()
	}

	// 开启用户从客户端读取数据流程的Goroutine
	routine.Go(w.ctx, func(ctx context.Context) error {
		w.StartReader()
		return nil
	})

	select {
	case <-w.ctx.Done():
		w.finalizer()
		return
	}
}

func (w *WsConnection) Stop() {
	w.cancel()
}

func (w *WsConnection) Context() context.Context {
	return w.ctx
}

func (w *WsConnection) GetName() string {
	return w.name
}

func (w *WsConnection) GetConnection() net.Conn {
	return nil
}

func (w *WsConnection) GetWsConn() *websocket.Conn {
	return w.conn
}

func (w *WsConnection) GetConnID() uint64 {
	return w.connID
}

func (w *WsConnection) GetMsgHandler() iface.IMsgHandle {
	return w.msgHandler
}

func (w *WsConnection) RemoteAddr() net.Addr {
	return w.conn.RemoteAddr()
}

func (w *WsConnection) LocalAddr() net.Addr {
	return w.conn.LocalAddr()
}

func (w *WsConnection) LocalAddrString() string {
	return w.localAddr
}

func (w *WsConnection) RemoteAddrString() string {
	return w.remoteAddr
}

func (w *WsConnection) Send(data []byte) error {
	w.msgLock.RLock()
	defer w.msgLock.RUnlock()
	if w.isClosed == true {
		return berr.ErrWsConnectionClose
	}

	err := w.conn.WriteMessage(websocket.BinaryMessage, data)
	if err != nil {
		log.Errorf("【BaiX】SendMsg err data = %+v, err = %+v", data, err)
		return err
	}
	return nil
}

func (w *WsConnection) SendMsg(msgID uint32, data []byte) error {
	w.msgLock.RLock()
	defer w.msgLock.RUnlock()
	if w.isClosed == true {
		return berr.ErrWsConnectionClose
	}

	// 将data封包，并且发送
	msg, err := w.packet.Pack(pack.NewMsgPackage(msgID, data))
	if err != nil {
		log.Errorf("【BaiX】Pack error msg ID = %d", msgID)
		return berr.ErrPackFail
	}

	err = w.conn.WriteMessage(websocket.BinaryMessage, msg)
	if err != nil {
		log.Errorf("【BaiX】SendMsg err msg ID = %d, data = %+v, err = %+v", msgID, string(msg), err)
		return err
	}
	return nil
}

func (w *WsConnection) SendBuffMsg(msgID uint32, data []byte) error {
	w.msgLock.RLock()
	defer w.msgLock.RUnlock()

	if w.msgBuffChan == nil {
		w.msgBuffChan = make(chan []byte, conf.GlobalConfig.MaxMsgChanLen)
		// 开启用于写回客户端数据流程的Goroutine
		// 此方法只读取MsgBuffChan中的数据没调用SendBuffMsg可以分配内存和启用协程
		go w.StartWriter()
	}

	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if w.isClosed == true {
		log.Errorf("【BaiX】err is %v", berr.ErrWsConnectionClose)
		return berr.ErrWsConnectionClose
	}

	// 将data封包，并且发送
	msg, err := w.packet.Pack(pack.NewMsgPackage(msgID, data))
	if err != nil {
		log.Errorf("【BaiX】Pack error msg ID = %d", msgID)
		return berr.ErrPackFail
	}

	select {
	case <-idleTimeout.C:
		return berr.ErrSendBuffMsgTimeOut
	case w.msgBuffChan <- msg:
		return nil
	}
}

func (w *WsConnection) SetProperty(key string, value interface{}) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()
	if w.property == nil {
		w.property = make(map[string]interface{})
	}

	w.property[key] = value
}

func (w *WsConnection) GetProperty(key string) (interface{}, error) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()

	if value, ok := w.property[key]; ok {
		return value, nil
	}

	return nil, berr.ErrNoPropertyFound
}

func (w *WsConnection) RemoveProperty(key string) {
	w.propertyLock.Lock()
	defer w.propertyLock.Unlock()

	delete(w.property, key)
}

func (w *WsConnection) IsAlive() bool {
	if w.isClosed {
		return false
	}
	// 检查连接最后一次活动时间，如果超过心跳间隔，则认为连接已经死亡
	return time.Now().Sub(w.lastActivityTime) < conf.GlobalConfig.HeartbeatMaxDuration()
}

func (w *WsConnection) SetHeartBeat(checker iface.IHeartbeatChecker) {
	w.hc = checker
}
