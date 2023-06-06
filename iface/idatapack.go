package iface

/*
封包数据和拆包数据
直接面向TCP连接中的数据流,为传输数据添加头部信息，用于处理TCP粘包问题。
*/

type IDataPack interface {
	GetHeadLen() uint32                // 获取包头长度方法
	Pack(msg IMessage) ([]byte, error) // 封包方法
	Unpack([]byte) (IMessage, error)   // 拆包方法
}

const (
	BaiXDataPack string = "baix_pack_tlv_big_endian" // BaiX 标准封包和拆包方式

	//...(+)
	// 自定义封包方式在此添加
)

const (
	BaiXMessage string = "baix_message" // BaiX 默认标准报文协议格式
)
