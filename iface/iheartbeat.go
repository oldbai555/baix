package iface

// 心跳检测

type IHeartbeatChecker interface {
	SetOnRemoteNotAliveFunc(OnRemoteNotAlive)
	SetHeartbeatMsgFunc(HeartBeatMsgFunc)
	SetHeartbeatFunc(HeartBeatFunc)
	BindRouter(msgID uint32, router IRouter)
	Start()
	Stop()
	SendHeartBeatMsg() error
	BindConn(IConnection)
	Clone() IHeartbeatChecker
	MsgID() uint32
	Router() IRouter
}

type HeartBeatMsgFunc func(IConnection) []byte // 用户自定义的心跳检测消息处理方法

type HeartBeatFunc func(IConnection) error // 用户自定义心跳函数

type OnRemoteNotAlive func(IConnection) // 用户自定义的远程连接不存活时的处理方法

const (
	HeartBeatDefaultMsgID uint32 = 99999
)
