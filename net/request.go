package net

import (
	"github.com/oldbai555/baix/iface"
	"sync"
)

var _ iface.IRequest = (*Request)(nil)

type Request struct {
}

func NewRequest(conn iface.IConnection, msg iface.IMessage) iface.IRequest {
	req := new(Request)
	req.steps = PRE_HANDLE
	req.conn = conn
	req.msg = msg
	req.stepLock = new(sync.RWMutex)
	req.needNext = true
	req.index = -1
	return req
}

func (r *Request) GetConnection() iface.IConnection {
	//TODO implement me
	panic("implement me")
}

func (r *Request) GetData() []byte {
	//TODO implement me
	panic("implement me")
}

func (r *Request) GetMsgID() uint32 {
	//TODO implement me
	panic("implement me")
}

func (r *Request) GetMessage() iface.IMessage {
	//TODO implement me
	panic("implement me")
}

func (r *Request) GetResponse() iface.IcResp {
	//TODO implement me
	panic("implement me")
}

func (r *Request) SetResponse(resp iface.IcResp) {
	//TODO implement me
	panic("implement me")
}

func (r *Request) BindRouter(router iface.IRouter) {
	//TODO implement me
	panic("implement me")
}

func (r *Request) Call() {
	//TODO implement me
	panic("implement me")
}

func (r *Request) Abort() {
	//TODO implement me
	panic("implement me")
}

func (r *Request) Goto(step iface.HandleStep) {
	//TODO implement me
	panic("implement me")
}

func (r *Request) BindRouterSlices(handlers []iface.RouterHandler) {
	//TODO implement me
	panic("implement me")
}

func (r *Request) RouterSlicesNext() {
	//TODO implement me
	panic("implement me")
}
