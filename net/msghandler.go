package net

import "github.com/oldbai555/baix/iface"

var _ iface.IMsgHandle = (*MsgHandler)(nil)

type MsgHandler struct{}

func (m *MsgHandler) AddRouter(msgID uint32, router iface.IRouter) {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) AddRouterSlices(msgID uint32, handler ...iface.RouterHandler) iface.IRouterSlices {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) Group(start, end uint32, handlers ...iface.RouterHandler) iface.IGroupRouterSlices {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) Use(Handlers ...iface.RouterHandler) iface.IRouterSlices {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) StartWorkerPool() {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) SendMsgToTaskQueue(request iface.IRequest) {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) Execute(request iface.IRequest) {
	//TODO implement me
	panic("implement me")
}

func (m *MsgHandler) AddInterceptor(interceptor iface.IInterceptor) {
	//TODO implement me
	panic("implement me")
}
