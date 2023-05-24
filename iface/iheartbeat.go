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

// HeartBeatMsgFunc User-defined method for handling heartbeat detection messages
// (用户自定义的心跳检测消息处理方法)
type HeartBeatMsgFunc func(IConnection) []byte

// HeartBeatFunc User-defined heartbeat function
// (用户自定义心跳函数)
type HeartBeatFunc func(IConnection) error

// OnRemoteNotAlive User-defined method for handling remote connections that are not alive
// 用户自定义的远程连接不存活时的处理方法
type OnRemoteNotAlive func(IConnection)

type HeartBeatOption struct {
	MakeMsg          HeartBeatMsgFunc // User-defined method for handling heartbeat detection messages(用户自定义的心跳检测消息处理方法)
	OnRemoteNotAlive OnRemoteNotAlive // User-defined method for handling remote connections that are not alive(用户自定义的远程连接不存活时的处理方法)
	HeadBeatMsgID    uint32           // User-defined ID for heartbeat detection messages(用户自定义的心跳检测消息ID)
	Router           IRouter          // User-defined business processing route for heartbeat detection messages(用户自定义的心跳检测消息业务处理路由)
}

const (
	HeartBeatDefaultMsgID uint32 = 99999
)
