package iface

import (
	"context"
	"net"

	"github.com/gorilla/websocket"
)

// IConnection Define connection interface
type IConnection interface {
	// Start the connection, make the current connection start working
	// (启动连接，让当前连接开始工作)
	Start()
	// Stop the connection and end the current connection state
	// (停止连接，结束当前连接状态)
	Stop()

	// Context Returns ctx, used by user-defined go routines to obtain connection exit status
	// (返回ctx，用于用户自定义的go程获取连接退出状态)
	Context() context.Context

	GetName() string            // Get the current connection name (获取当前连接名称)
	GetConnection() net.Conn    // Get the original socket from the current connection(从当前连接获取原始的socket)
	GetWsConn() *websocket.Conn // Get the original websocket connection from the current connection(从当前连接中获取原始的websocket连接)
	GetConnID() uint64          // Get the current connection ID (获取当前连接ID)
	GetMsgHandler() IMsgHandle  // Get the message handler (获取消息处理器)
	RemoteAddr() net.Addr       // Get the remote address information of the connection (获取链接远程地址信息)
	LocalAddr() net.Addr        // Get the local address information of the connection (获取链接本地地址信息)
	LocalAddrString() string    // Get the local address information of the connection as a string
	RemoteAddrString() string   // Get the remote address information of the connection as a string

	Send(data []byte) error // Send data directly to the remote TCP client (without buffering)

	SendMsg(msgId uint32, data []byte) error // 直接将Message数据发送数据给远程的TCP客户端(无缓冲)

	// SendBuffMsg Send Message data to the message queue to be sent to the remote TCP client later (with buffering)
	// 直接将Message数据发送给远程的TCP客户端(有缓冲)
	SendBuffMsg(msgId uint32, data []byte) error

	SetProperty(key string, value interface{})   // Set connection property
	GetProperty(key string) (interface{}, error) // Get connection property
	RemoveProperty(key string)                   // Remove connection property
	IsAlive() bool                               // Check if the current connection is alive(判断当前连接是否存活)
	SetHeartBeat(checker IHeartbeatChecker)      // Set the heartbeat detector (设置心跳检测器)
}
