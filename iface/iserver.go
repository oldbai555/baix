package iface

import (
	"net/http"
	"time"
)

type IServer interface {
	Start() // Start 启动服务器方法
	Stop()  // Stop 停止服务器方法
	Serve() // Serve 开启业务服务方法

	// AddRouter 路由功能：给当前服务注册一个路由业务方法，供客户端链接处理使用
	AddRouter(msgID uint32, router IRouter)

	// GetConnMgr 得到链接管理
	GetConnMgr() IConnManager

	// SetOnConnStart 设置该Server的连接创建时Hook函数
	SetOnConnStart(func(IConnection))

	// SetOnConnStop 设置该Server的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))

	// GetOnConnStart 得到该Server的连接创建时Hook函数
	GetOnConnStart() func(IConnection)

	// GetOnConnStop 得到该Server的连接断开时的Hook函数
	GetOnConnStop() func(IConnection)

	// GetPacket 获取Server绑定的数据协议封包方式
	GetPacket() IDataPack

	// GetMsgHandler 获取Server绑定的消息处理模块
	GetMsgHandler() IMsgHandle

	// SetPacket 设置Server绑定的数据协议封包方式
	SetPacket(IDataPack)

	// StartHeartBeat 启动心跳检测
	StartHeartBeat(time.Duration)

	// GetHeartBeat 获取心跳检测器
	GetHeartBeat() IHeartbeatChecker

	GetLengthField() *LengthField
	SetDecoder(IDecoder)
	AddInterceptor(IInterceptor)

	// SetWebsocketAuth 添加websocket认证方法
	SetWebsocketAuth(func(r *http.Request) error)

	// ServerName 获取服务器名称
	ServerName() string
}
