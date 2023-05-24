package net

import "github.com/oldbai555/lbtool/pkg/baix/iface"

var _ iface.IConnManager = (*ConnManager)(nil)

type ConnManager struct{}

func (c *ConnManager) Add(conn iface.IConnection) {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) Remove(conn iface.IConnection) {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) Get(connId string) (iface.IConnection, error) {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) Len() int {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) ClearAllConn() {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) HasIdConn(connId string) bool {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) NotifyBuffToConnByID(connId string, m iface.IMessage) error {
	//TODO implement me
	panic("implement me")
}

func (c *ConnManager) NotifyBuffAll(m iface.IMessage) error {
	//TODO implement me
	panic("implement me")
}
