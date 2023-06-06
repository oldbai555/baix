package conf

import (
	"github.com/oldbai555/lbtool/log"
	"reflect"
	"time"
)

const (
	ServerModeTcp       = "tcp"
	ServerModeWebsocket = "websocket"
)

type Config struct {
	Host    string // 当前服务器主机IP
	TCPPort int    // 当前服务器主机监听端口号
	WsPort  int    // 当前服务器主机websocket监听端口
	Name    string // 当前服务器名称

	Version          string // 当前 BaiX 版本号
	MaxPacketSize    uint32 // 读写数据包的最大值
	MaxConn          int    // 当前服务器主机允许的最大链接个数
	WorkerPoolSize   uint32 // 业务工作Worker池的数量
	MaxWorkerTaskLen uint32 // 业务工作Worker对应负责的任务队列最大任务存储数量
	MaxMsgChanLen    uint32 // SendBuffMsg发送消息的缓冲最大长度
	IOReadBuffSize   uint32 // 每次IO最大的读取长度

	Mode string // "tcp":tcp监听, "websocket":websocket 监听 为空时同时开启

	HeartbeatMax int // 最长心跳检测间隔时间(单位：秒),超过改时间间隔，则认为超时，从配置文件读取

	/*
		TLS
	*/
	CertFile       string // 证书文件名称 默认""
	PrivateKeyFile string // 私钥文件名称 默认"" --如果没有设置证书和私钥文件，则不启用TLS加密
}

var GlobalConfig *Config // 定义一个全局的对象

func (g *Config) Show() {
	objVal := reflect.ValueOf(g).Elem()
	objType := reflect.TypeOf(*g)

	log.Infof("===== 【BaiX】 Global Config =====")
	for i := 0; i < objVal.NumField(); i++ {
		field := objVal.Field(i)
		typeField := objType.Field(i)

		log.Infof("%s: %v", typeField.Name, field.Interface())
	}
	log.Infof("==============================")
}

func (g *Config) HeartbeatMaxDuration() time.Duration {
	return time.Duration(g.HeartbeatMax) * time.Second
}

func init() {
	// 初始化GlobalObject变量，设置一些默认值
	GlobalConfig = &Config{
		Host:             "127.0.0.1",
		TCPPort:          8999,
		WsPort:           9000,
		Name:             "【BaiX】ServerApp",
		Version:          "V1.0",
		MaxPacketSize:    4096,
		MaxConn:          12000,
		WorkerPoolSize:   1,
		MaxWorkerTaskLen: 1024,
		MaxMsgChanLen:    1024,
		IOReadBuffSize:   1024,
		Mode:             ServerModeTcp,
		HeartbeatMax:     10, // 默认心跳检测最长间隔为10秒
	}
}
