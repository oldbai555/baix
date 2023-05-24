package conf

import (
	"fmt"
	"reflect"
	"time"
)

const (
	ServerModeTcp       = "tcp"
	ServerModeWebsocket = "websocket"
)

type Config struct {
	Host    string //The IP address of the current server. (当前服务器主机IP)
	TCPPort int    //The port number on which the server listens for TCP connections.(当前服务器主机监听端口号)
	WsPort  int    //The port number on which the server listens for WebSocket connections.(当前服务器主机websocket监听端口)
	Name    string //The name of the current server.(当前服务器名称)

	Version          string //The version of the BaiX framework.(当前 BaiX 版本号)
	MaxPacketSize    uint32 //The maximum size of the packets that can be sent or received.(读写数据包的最大值)
	MaxConn          int    //The maximum number of connections that the server can handle.(当前服务器主机允许的最大链接个数)
	WorkerPoolSize   uint32 //The number of worker pools in the business logic.(业务工作Worker池的数量)
	MaxWorkerTaskLen uint32 //The maximum number of tasks that a worker pool can handle.(业务工作Worker对应负责的任务队列最大任务存储数量)
	MaxMsgChanLen    uint32 //The maximum length of the send buffer message queue.(SendBuffMsg发送消息的缓冲最大长度)
	IOReadBuffSize   uint32 //The maximum size of the read buffer for each IO operation.(每次IO最大的读取长度)

	Mode string //"tcp":tcp监听, "websocket":websocket 监听 为空时同时开启

	HeartbeatMax int // 最长心跳检测间隔时间(单位：秒),超过改时间间隔，则认为超时，从配置文件读取

	/*
		TLS
	*/
	CertFile       string // The name of the certificate file. If it is empty, TLS encryption is not enabled.(证书文件名称 默认"")
	PrivateKeyFile string // The name of the private key file. If it is empty, TLS encryption is not enabled.(私钥文件名称 默认"" --如果没有设置证书和私钥文件，则不启用TLS加密)
}

var GlobalConfig *Config // Define a global object.(定义一个全局的对象)

func (g *Config) Show() {
	objVal := reflect.ValueOf(g).Elem()
	objType := reflect.TypeOf(*g)

	fmt.Println("===== Baix Global Config =====")
	for i := 0; i < objVal.NumField(); i++ {
		field := objVal.Field(i)
		typeField := objType.Field(i)

		fmt.Printf("%s: %v\n", typeField.Name, field.Interface())
	}
	fmt.Println("==============================")
}

func (g *Config) HeartbeatMaxDuration() time.Duration {
	return time.Duration(g.HeartbeatMax) * time.Second
}

func init() {
	// 初始化GlobalObject变量，设置一些默认值
	GlobalConfig = &Config{
		Host:             "0.0.0.0",
		TCPPort:          8999,
		WsPort:           9000,
		Name:             "BaiXServerApp",
		Version:          "V1.0",
		MaxPacketSize:    4096,
		MaxConn:          12000,
		WorkerPoolSize:   10,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
		IOReadBuffSize:   1024,
		Mode:             ServerModeTcp,
		HeartbeatMax:     10, // 默认心跳检测最长间隔为10秒
	}
}
