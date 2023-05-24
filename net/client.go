package net

import (
	"github.com/oldbai555/baix/iface"
	"time"
)

var _ iface.IClient = (*Client)(nil)

type Client struct{}

func (c *Client) Restart() {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Start() {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Stop() {
	//TODO implement me
	panic("implement me")
}

func (c *Client) AddRouter(msgID uint32, router iface.IRouter) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) Conn() iface.IConnection {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetOnConnStart(f func(iface.IConnection)) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetOnConnStop(f func(iface.IConnection)) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetOnConnStart() func(iface.IConnection) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetOnConnStop() func(iface.IConnection) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetPacket() iface.IDataPack {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetPacket(pack iface.IDataPack) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetMsgHandler() iface.IMsgHandle {
	//TODO implement me
	panic("implement me")
}

func (c *Client) StartHeartBeat(duration time.Duration) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) StartHeartBeatWithOption(duration time.Duration, option *iface.HeartBeatOption) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetLengthField() *iface.LengthField {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetDecoder(decoder iface.IDecoder) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) AddInterceptor(interceptor iface.IInterceptor) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetErrChan() chan error {
	//TODO implement me
	panic("implement me")
}

func (c *Client) SetName(s string) {
	//TODO implement me
	panic("implement me")
}

func (c *Client) GetName() string {
	//TODO implement me
	panic("implement me")
}
