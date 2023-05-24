package iface

import (
	"net/http"
	"time"
)

// IServer Defines the server interface
type IServer interface {
	Start() // Start the server method(启动服务器方法)
	Stop()  // Stop the server method (停止服务器方法)
	Serve() // Start the business service method(开启业务服务方法)

	// AddRouter Routing feature: register a routing business method for the current service for client link processing use
	//(路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用)
	AddRouter(msgID uint32, router IRouter)

	// GetConnMgr Get connection management (得到链接管理)
	GetConnMgr() IConnManager

	// SetOnConnStart Set Hook function when the connection is created for the Server (设置该Server的连接创建时Hook函数)
	SetOnConnStart(func(IConnection))

	// SetOnConnStop Set Hook function when the connection is disconnected for the Server
	// (设置该Server的连接断开时的Hook函数)
	SetOnConnStop(func(IConnection))

	// GetOnConnStart Get Hook function when the connection is created for the Server
	// (得到该Server的连接创建时Hook函数)
	GetOnConnStart() func(IConnection)

	// GetOnConnStop Get Hook function when the connection is disconnected for the Server
	// (得到该Server的连接断开时的Hook函数)
	GetOnConnStop() func(IConnection)

	// GetPacket Get the data protocol packet binding method for the Server
	// (获取Server绑定的数据协议封包方式)
	GetPacket() IDataPack

	// GetMsgHandler Get the message processing module binding method for the Server
	// (获取Server绑定的消息处理模块)
	GetMsgHandler() IMsgHandle

	// SetPacket Set the data protocol packet binding method for the Server
	// (设置Server绑定的数据协议封包方式)
	SetPacket(IDataPack)

	// StartHeartBeat Start the heartbeat check
	// (启动心跳检测)
	StartHeartBeat(time.Duration)

	// GetHeartBeat Get the heartbeat checker
	// (获取心跳检测器)
	GetHeartBeat() IHeartbeatChecker

	GetLengthField() *LengthField
	SetDecoder(IDecoder)
	AddInterceptor(IInterceptor)

	// SetWebsocketAuth Add WebSocket authentication method
	// (添加websocket认证方法)
	SetWebsocketAuth(func(r *http.Request) error)

	// ServerName Get the server name (获取服务器名称)
	ServerName() string
}
