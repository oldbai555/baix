package net

import (
	"context"
	"github.com/gorilla/websocket"
	"github.com/oldbai555/baix/iface"
	"net"
)

var _ iface.IConnection = (*WsConnection)(nil)

type WsConnection struct{}

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
		c.frameDecoder = zinterceptor.NewFrameDecoder(*lengthField)
	}

	// Inherited attributes from server (从server继承过来的属性)
	c.packet = server.GetPacket()
	c.onConnStart = server.GetOnConnStart()
	c.onConnStop = server.GetOnConnStop()
	c.msgHandler = server.GetMsgHandler()

	// Bind the current Connection to the Server's ConnManager (将当前的Connection与Server的ConnManager绑定)
	c.connManager = server.GetConnMgr()

	// Add the newly created Conn to the connection management (将新创建的Conn添加到链接管理中)
	server.GetConnMgr().Add(c)

	return c
}

func (w *WsConnection) Start() {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) Stop() {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) Context() context.Context {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetName() string {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetConnection() net.Conn {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetWsConn() *websocket.Conn {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetConnID() uint64 {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetMsgHandler() iface.IMsgHandle {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) LocalAddrString() string {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) RemoteAddrString() string {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) Send(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) SendToQueue(data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) SendMsg(msgId uint32, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) SendBuffMsg(msgId uint32, data []byte) error {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) SetProperty(key string, value interface{}) {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) GetProperty(key string) (interface{}, error) {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) RemoveProperty(key string) {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) IsAlive() bool {
	//TODO implement me
	panic("implement me")
}

func (w *WsConnection) SetHeartBeat(checker iface.IHeartbeatChecker) {
	//TODO implement me
	panic("implement me")
}
