package net

import (
	"github.com/oldbai555/baix/berr"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/lbtool/log"
	"sync"
)

var _ iface.IConnManager = (*ConnManager)(nil)

type ConnManager struct {
	connections map[uint64]iface.IConnection
	connLock    sync.RWMutex
}

func newConnManager() iface.IConnManager {
	return &ConnManager{
		connections: make(map[uint64]iface.IConnection),
	}
}

func (c *ConnManager) Add(connection iface.IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	c.connections[connection.GetConnID()] = connection
	log.Infof("【BaiX】connection add to ConnManager successfully: conn num = %d", c.Len())
}

func (c *ConnManager) Remove(connection iface.IConnection) {
	c.connLock.Lock()
	defer c.connLock.Unlock()
	delete(c.connections, connection.GetConnID())
	log.Infof("【BaiX】connection Remove ConnID=%d successfully: conn num = %d", connection.GetConnID(), c.Len())
}

func (c *ConnManager) Get(u uint64) (iface.IConnection, error) {
	c.connLock.RLock()
	defer c.connLock.RUnlock()

	if conn, ok := c.connections[u]; ok {
		return conn, nil
	}

	return nil, berr.ErrConnNotFound
}

func (c *ConnManager) Len() int {
	length := len(c.connections)
	return length
}

func (c *ConnManager) ClearConn() {
	c.connLock.Lock()
	defer c.connLock.Unlock()

	for connID, conn := range c.connections {
		conn.Stop()
		delete(c.connections, connID)
	}

	log.Infof("【BaiX】Clear All Connections successfully: conn num = %d", c.Len())
}

func (c *ConnManager) GetAllConnID() []uint64 {
	ids := make([]uint64, 0, len(c.connections))

	c.connLock.RLock()
	defer c.connLock.RUnlock()

	for id := range c.connections {
		ids = append(ids, id)
	}

	return ids
}

func (c *ConnManager) Range(cb func(uint64, iface.IConnection, interface{}) error, args interface{}) (err error) {
	for connID, conn := range c.connections {
		err = cb(connID, conn, args)
	}
	return
}
