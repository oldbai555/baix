package tcp

import (
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/net"
	"github.com/oldbai555/lbtool/log"
	"os"
	"os/signal"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	client := net.NewClient(conf.GlobalConfig.Host, conf.GlobalConfig.TCPPort)
	client.StartHeartBeat(time.Second)
	client.Start()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("===exit===", sig)
	client.Stop()
	time.Sleep(time.Second * 2)
}

func TestNewClient1(t *testing.T) {
	client := net.NewClient(conf.GlobalConfig.Host, conf.GlobalConfig.TCPPort)
	client.StartHeartBeat(time.Second)
	client.Start()

	// close
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	sig := <-c
	log.Infof("===exit===", sig)
	client.Stop()
	time.Sleep(time.Second * 2)
}
