package iface

import (
	"context"
	"net"

	"github.com/gorilla/websocket"
)

// IConnection Define connection interface
type IConnection interface {
	Start()                   // 启动连接，让当前连接开始工作
	Stop()                    // 停止连接，结束当前连接状态
	Context() context.Context // 返回ctx，用于用户自定义的go程获取连接退出状态

	GetName() string            // 获取当前连接名称
	GetConnection() net.Conn    // 从当前连接获取原始的socket
	GetWsConn() *websocket.Conn // 从当前连接中获取原始的websocket连接
	GetConnID() uint64          // 获取当前连接ID
	GetMsgHandler() IMsgHandle  // 获取消息处理器
	RemoteAddr() net.Addr       // 获取链接远程地址信息
	LocalAddr() net.Addr        // 获取链接本地地址信息
	LocalAddrString() string
	RemoteAddrString() string

	Send(data []byte) error // 将数据直接发送到远程TCP客户端(无缓冲)

	SendMsg(msgId uint32, data []byte) error // 直接将Message数据发送数据给远程的TCP客户端(无缓冲)

	SendBuffMsg(msgId uint32, data []byte) error // 直接将Message数据发送给远程的TCP客户端(有缓冲)

	SetProperty(key string, value interface{})   // Set connection property
	GetProperty(key string) (interface{}, error) // Get connection property
	RemoveProperty(key string)                   // Remove connection property
	IsAlive() bool                               // 判断当前连接是否存活
	SetHeartBeat(checker IHeartbeatChecker)      // 设置心跳检测器
}
