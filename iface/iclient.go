package iface

import "time"

type IClient interface {
	Restart()
	Start()
	Stop()
	AddRouter(msgID uint32, router IRouter)
	Conn() IConnection

	// SetOnConnStart 设置该Client的连接创建时Hook函数
	SetOnConnStart(func(IConnection))

	// SetOnConnStop 设置该Client的连接断开时的Hook函数
	SetOnConnStop(func(IConnection))

	// GetOnConnStart 获取该Client的连接创建时Hook函数
	GetOnConnStart() func(IConnection)

	// GetOnConnStop 设置该Client的连接断开时的Hook函数
	GetOnConnStop() func(IConnection)

	// GetPacket 获取Client绑定的数据协议封包方式
	GetPacket() IDataPack

	// SetPacket 设置Client绑定的数据协议封包方式
	SetPacket(IDataPack)

	// GetMsgHandler 获取Client绑定的消息处理模块
	GetMsgHandler() IMsgHandle

	// StartHeartBeat 启动心跳检测
	StartHeartBeat(time.Duration)

	// GetLengthField 获取长度字段
	GetLengthField() *LengthField

	// SetDecoder 设置解码器
	SetDecoder(IDecoder)

	// AddInterceptor 添加拦截器
	AddInterceptor(IInterceptor)

	// GetErrChan 获取客户端错误管道
	GetErrChan() chan error

	// SetName 设置客户端Client名称
	SetName(string)

	// GetName 获取客户端Client名称
	GetName() string
}
