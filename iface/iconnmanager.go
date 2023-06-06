package iface

// Connection Management Abstract Layer

type IConnManager interface {
	Add(IConnection)
	Remove(IConnection)
	Get(uint64) (IConnection, error)
	Len() int
	ClearConn()
	GetAllConnID() []uint64
	Range(func(uint64, IConnection, interface{}) error, interface{}) error // 遍历所有连接
}
