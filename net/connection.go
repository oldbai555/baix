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

var _ iface.IConnection = (*Connection)(nil)

type Connection struct {
	// // The socket TCP socket of the current connection(当前连接的socket TCP套接字)
	conn net.Conn

	// The ID of the current connection, also known as SessionID, globally unique, used by server Connection
	// uint64 range: 0~18,446,744,073,709,551,615
	// This is the maximum number of connID theoretically supported by the process
	// (当前连接的ID 也可以称作为SessionID，ID全局唯一 ，服务端Connection使用
	// uint64 取值范围：0 ~ 18,446,744,073,709,551,615
	// 这个是理论支持的进程connID的最大数量)
	connID uint64

	// The message management module that manages MsgID and the corresponding processing method
	// (消息管理MsgID和对应处理方法的消息管理模块)
	msgHandler iface.IMsgHandle

	// Channel to notify that the connection has exited/stopped
	// (告知该链接已经退出/停止的channel)
	ctx    context.Context
	cancel context.CancelFunc

	// Buffered channel used for message communication between the read and write goroutines
	// (有缓冲管道，用于读、写两个goroutine之间的消息通信)
	msgBuffChan chan []byte

	// Lock for user message reception and transmission
	// (用户收发消息的Lock)
	msgLock sync.RWMutex

	// Connection properties
	// (链接属性)
	property map[string]interface{}

	// Lock to protect the current property
	// (保护当前property的锁)
	propertyLock sync.Mutex

	// The current connection's close state
	// (当前连接的关闭状态)
	isClosed bool

	// Which Connection Manager the current connection belongs to
	// (当前链接是属于哪个Connection Manager的)
	connManager iface.IConnManager

	// Hook function when the current connection is created
	// (当前连接创建时Hook函数)
	onConnStart func(conn iface.IConnection)

	// Hook function when the current connection is disconnected
	// (当前连接断开时的Hook函数)
	onConnStop func(conn iface.IConnection)

	// Data packet packaging method
	// (数据报文封包方式)
	packet iface.IDataPack

	// Last activity time
	// (最后一次活动时间)
	lastActivityTime time.Time

	// Framedecoder for solving fragmentation and packet sticking problems
	// (断粘包解码器)
	frameDecoder iface.IFrameDecoder

	// Heartbeat checker
	// (心跳检测器)
	hc iface.IHeartbeatChecker

	// Connection name, default to be the same as the name of the Server/Client that created the connection
	// (链接名称，默认与创建链接的Server/Client的Name一致)
	name string

	// Local address of the current connection
	// (当前链接的本地地址)
	localAddr string

	// Remote address of the current connection
	// (当前链接的远程地址)
	remoteAddr string
}

// newServerConn 创建一个Server服务端特性的连接的方法
func newServerConn(server iface.IServer, conn net.Conn, connID uint64) iface.IConnection {

	// Initialize Conn properties
	c := &Connection{
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

	// Inherited properties from server (从server继承过来的属性)
	c.packet = server.GetPacket()
	c.onConnStart = server.GetOnConnStart()
	c.onConnStop = server.GetOnConnStop()
	c.msgHandler = server.GetMsgHandler()

	// Bind the current Connection with the Server's ConnManager
	// (将当前的Connection与Server的ConnManager绑定)
	c.connManager = server.GetConnMgr()

	// Add the newly created Conn to the connection manager
	// (将新创建的Conn添加到链接管理中)
	server.GetConnMgr().Add(c)

	return c
}

func (c *Connection) callOnConnStart() {
	if c.onConnStart != nil {
		log.Infof("BaiX CallOnConnStart....")
		c.onConnStart(c)
	}
}

func (c *Connection) callOnConnStop() {
	if c.onConnStop != nil {
		log.Infof("BaiX CallOnConnStop....")
		c.onConnStop(c)
	}
}

func (c *Connection) updateActivity() {
	c.lastActivityTime = time.Now()
}

func (c *Connection) startReader() {
	log.Infof("[Reader Goroutine is running]")
	defer func() {
		log.Infof("%s [conn Reader exit!]", c.RemoteAddr().String())
		c.Stop()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			buffer := make([]byte, conf.GlobalConfig.IOReadBuffSize)

			// 从conn的IO中读取数据到内存缓冲buffer中
			n, err := c.conn.Read(buffer)
			if err != nil {
				log.Errorf("read msg head [read datalen=%d], error = %s", n, err)
				return
			}
			log.Debugf("read buffer %s \n", hex.EncodeToString(buffer[0:n]))

			// 正常读取到对端数据，更新心跳检测Active状态
			if n > 0 && c.hc != nil {
				c.updateActivity()
			}

			// 处理自定义协议断粘包问题
			if c.frameDecoder != nil {
				// 为读取到的0-n个字节的数据进行解码
				bufArrays := c.frameDecoder.Decode(buffer[0:n])
				if bufArrays == nil {
					continue
				}
				for _, bytes := range bufArrays {
					msg := pack.NewMessage(uint32(len(bytes)), bytes)

					// 得到当前客户端请求的Request数据
					req := NewRequest(c, msg)
					c.msgHandler.Execute(req)
				}
				continue
			}

			// 正常解码
			msg := pack.NewMessage(uint32(n), buffer[0:n])

			// Get the current client's Request data
			// (得到当前客户端请求的Request数据)
			req := NewRequest(c, msg)
			c.msgHandler.Execute(req)
		}
	}
}

// 写消息Goroutine， 用户将数据发送给客户端
func (c *Connection) startWriter() {
	log.Infof("Writer Goroutine is running")
	defer func() {
		log.Infof("%s [conn Writer exit!]", c.RemoteAddr().String())
	}()

	for {
		select {
		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.conn.Write(data); err != nil {
					log.Errorf("Send Buff Data error:, %s Conn Writer exit", err)
					break
				}

			} else {
				log.Errorf("msgBuffChan is Closed")
				break
			}
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *Connection) finalizer() {
	// Call the callback function registered by the user when closing the connection if it exists
	// (如果用户注册了该链接的	关闭回调业务，那么在此刻应该显示调用)
	c.callOnConnStop()

	c.msgLock.Lock()
	defer c.msgLock.Unlock()

	// If the connection has already been closed
	if c.isClosed == true {
		return
	}

	// Stop the heartbeat detector associated with the connection
	if c.hc != nil {
		c.hc.Stop()
	}

	// Close the socket connection
	_ = c.conn.Close()

	// Remove the connection from the connection manager
	if c.connManager != nil {
		c.connManager.Remove(c)
	}

	// Close all channels associated with the connection
	if c.msgBuffChan != nil {
		close(c.msgBuffChan)
	}

	c.isClosed = true

	log.Infof("Conn Stop()...ConnID = %d", c.connID)
}

func (c *Connection) Start() {
	c.ctx, c.cancel = context.WithCancel(context.Background())

	// Execute the hook method for processing business logic when creating a connection
	// (按照用户传递进来的创建连接时需要处理的业务，执行钩子方法)
	c.callOnConnStart()

	// Start heart beating detection
	if c.hc != nil {
		c.hc.Start()
		c.updateActivity()
	}

	// Start the Goroutine for reading data from the client
	// (开启用户从客户端读取数据流程的Goroutine)
	routine.Go(c.ctx, func(ctx context.Context) error {
		c.startReader()
		return nil
	})

	select {
	case <-c.ctx.Done():
		c.finalizer()
		return
	}
}

func (c *Connection) Stop() {
	c.cancel()
}

func (c *Connection) Context() context.Context {
	return c.ctx
}

func (c *Connection) GetName() string {
	return c.name
}

func (c *Connection) GetConnection() net.Conn {
	return c.conn
}

func (c *Connection) GetWsConn() *websocket.Conn {
	return nil
}

func (c *Connection) GetConnID() uint64 {
	return c.connID
}

func (c *Connection) GetMsgHandler() iface.IMsgHandle {
	return c.msgHandler
}

func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Connection) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *Connection) LocalAddrString() string {
	return c.localAddr
}

func (c *Connection) RemoteAddrString() string {
	return c.remoteAddr
}

func (c *Connection) Send(data []byte) error {
	c.msgLock.RLock()
	defer c.msgLock.RUnlock()
	if c.isClosed == true {
		log.Errorf("err is %v", berr.ErrConnectionClose)
		return berr.ErrConnectionClose
	}

	_, err := c.conn.Write(data)
	if err != nil {
		log.Errorf("SendMsg err data = %+v, err = %+v", data, err)
		return err
	}

	return nil
}

func (c *Connection) SendMsg(msgID uint32, data []byte) error {
	c.msgLock.RLock()
	defer c.msgLock.RUnlock()
	if c.isClosed == true {
		return berr.ErrConnectionClose
	}

	// Pack data and send it
	msg, err := c.packet.Pack(pack.NewMsgPackage(msgID, data))
	if err != nil {
		log.Errorf("Pack error msg ID = %d", msgID)
		return berr.ErrPackFail
	}

	_, err = c.conn.Write(msg)
	if err != nil {
		log.Errorf("SendMsg err msg ID = %d, data = %+v, err = %+v", msgID, string(msg), err)
		return err
	}
	return nil
}

func (c *Connection) SendBuffMsg(msgID uint32, data []byte) error {
	c.msgLock.RLock()
	defer c.msgLock.RUnlock()

	if c.msgBuffChan == nil {
		c.msgBuffChan = make(chan []byte, conf.GlobalConfig.MaxMsgChanLen)
		// 开启用于写回客户端数据流程的Goroutine
		// 此方法只读取MsgBuffChan中的数据没调用SendBuffMsg可以分配内存和启用协程
		routine.Go(c.ctx, func(ctx context.Context) error {
			c.startWriter()
			return nil
		})
	}

	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if c.isClosed == true {
		return berr.ErrConnectionClose
	}

	msg, err := c.packet.Pack(pack.NewMsgPackage(msgID, data))
	if err != nil {
		log.Errorf("Pack error msg ID = %d", msgID)
		return berr.ErrPackFail
	}

	// send timeout
	select {
	case <-idleTimeout.C:
		return berr.ErrSendBuffMsgTimeOut
	case c.msgBuffChan <- msg:
		return nil
	}
}

func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()
	if c.property == nil {
		c.property = make(map[string]interface{})
	}

	c.property[key] = value
}

func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	}

	return nil, berr.ErrNoPropertyFound
}

func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *Connection) IsAlive() bool {
	if c.isClosed {
		return false
	}
	// Check the last activity time of the connection. If it's beyond the heartbeat interval,
	// then the connection is considered dead.
	// (检查连接最后一次活动时间，如果超过心跳间隔，则认为连接已经死亡)
	return time.Now().Sub(c.lastActivityTime) < conf.GlobalConfig.HeartbeatMaxDuration()
}

func (c *Connection) SetHeartBeat(checker iface.IHeartbeatChecker) {
	c.hc = checker
}
