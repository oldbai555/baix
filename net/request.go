package net

import (
	"github.com/oldbai555/baix/iface"
	"sync"
)

const (
	PreHandle  iface.HandleStep = iota // PreHandle for pre-processing
	Handle                             // Handle for processing
	PostHandle                         // PostHandle for post-processing

	HandleOver
)

var _ iface.IRequest = (*Request)(nil)

type Request struct {
	iface.BaseRequest
	conn     iface.IConnection // 已经和客户端建立好的链接
	msg      iface.IMessage    // 客户端请求的数据
	router   iface.IRouter     // 请求处理的函数
	steps    iface.HandleStep  // 用来控制路由函数执行
	stepLock *sync.RWMutex     // 并发互斥
	needNext bool              // 是否需要执行下一个路由函数
	icResp   iface.IcResp      // 拦截器返回数据
	index    int8              // 路由函数切片索引
}

func NewRequest(conn iface.IConnection, msg iface.IMessage) iface.IRequest {
	req := new(Request)
	req.steps = PreHandle
	req.conn = conn
	req.msg = msg
	req.stepLock = new(sync.RWMutex)
	req.needNext = true
	req.index = -1
	return req
}

func (r *Request) next() {
	r.stepLock.Lock()
	defer r.stepLock.Unlock()
	if r.needNext == false {
		r.needNext = true
		return
	}
	r.steps++
}

func (r *Request) GetConnection() iface.IConnection {
	return r.conn
}

func (r *Request) GetData() []byte {
	return r.msg.GetData()
}

func (r *Request) GetMsgID() uint32 {
	return r.msg.GetMsgID()
}

func (r *Request) GetMessage() iface.IMessage {
	return r.msg
}

func (r *Request) GetResponse() iface.IcResp {
	return r.icResp
}

func (r *Request) SetResponse(resp iface.IcResp) {
	r.icResp = resp
}

func (r *Request) BindRouter(router iface.IRouter) {
	r.router = router
}

func (r *Request) Call() {
	if r.router == nil {
		return
	}

	for r.steps < HandleOver {
		switch r.steps {
		case PreHandle:
			r.router.PreHandle(r)
		case Handle:
			r.router.Handle(r)
		case PostHandle:
			r.router.PostHandle(r)
		}
		r.next()
	}

	r.steps = HandleOver
}

func (r *Request) Abort() {
	r.stepLock.Lock()
	defer r.stepLock.Unlock()
	r.steps = HandleOver
}
