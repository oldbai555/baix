package tcp

import (
	"github.com/oldbai555/baix/net"
	"github.com/oldbai555/lbtool/log"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	server := net.NewServer()
	server.StartHeartBeat(time.Second)
	server.Serve()
}

// 异或求重复数字
func TestDemo(t *testing.T) {
	var list = []int{1, 3, 2, 4, 4, 7, 6, 5, 8, 9, 10}
	val := list[0]
	log.Infof("val = %d", val)
	for i := 1; i < len(list); i++ {
		log.Infof("val is %d , list[%d]=%d ^ %d is %d", val, i, list[i], i, list[i]^i)
		val ^= list[i] ^ i // 表达式 a^b^b = a , a^b^a = a
		log.Infof("val = %d", val)
	}
	log.Infof("res = %d", val)
}
