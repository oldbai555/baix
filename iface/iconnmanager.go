package iface

// Connection Management Abstract Layer

type IConnManager interface {
	Add(IConnection)                                                       // Add connection
	Remove(IConnection)                                                    // Remove connection
	Get(uint64) (IConnection, error)                                       // Get a connection by ConnID
	Len() int                                                              // Get current number of connections
	ClearConn()                                                            // Remove and stop all connections
	GetAllConnID() []uint64                                                // Get all connection IDs
	Range(func(uint64, IConnection, interface{}) error, interface{}) error // Traverse all connections
}
