package iface

type HandleStep int

// IFuncRequest function message interface (函数消息接口)
type IFuncRequest interface {
	CallFunc()
}

/*
IRequest interface:

	It actually packages the connection information and request data of the client request into Request
	(实际上是把客户端请求的链接信息 和 请求的数据 包装到了 Request里)
*/
type IRequest interface {
	GetConnection() IConnection // Get the connection information of the request(获取请求连接信息)

	GetData() []byte  // Get the data of the request message(获取请求消息的数据)
	GetMsgID() uint32 // Get the message ID of the request(获取请求的消息ID)

	GetMessage() IMessage // Get the raw data of the request message (获取请求消息的原始数据 add by uuxia 2023-03-10)

	GetResponse() IcResp // Get the serialized data after parsing(获取解析完后序列化数据)
	SetResponse(IcResp)  // Set the serialized data after parsing(设置解析完后序列化数据)

	BindRouter(router IRouter) // Bind which router handles this request(绑定这次请求由哪个路由处理)

	// Call 转进到下一个处理器开始执行 但是调用此方法的函数会根据先后顺序逆序执行
	Call()

	// Abort 终止处理函数的运行 但调用此方法的函数会执行完毕
	Abort()
}

var _ IRequest = (*BaseRequest)(nil)

type BaseRequest struct{}

func (br *BaseRequest) Next()                      {}
func (br *BaseRequest) GetConnection() IConnection { return nil }
func (br *BaseRequest) GetData() []byte            { return nil }
func (br *BaseRequest) GetMsgID() uint32           { return 0 }
func (br *BaseRequest) GetMessage() IMessage       { return nil }
func (br *BaseRequest) GetResponse() IcResp        { return nil }
func (br *BaseRequest) SetResponse(resp IcResp)    {}
func (br *BaseRequest) BindRouter(router IRouter)  {}
func (br *BaseRequest) Call()                      {}
func (br *BaseRequest) Abort()                     {}
