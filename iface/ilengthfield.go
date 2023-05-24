package iface

import "encoding/binary"

// 帧解码器

type IFrameDecoder interface {
	Decode(buff []byte) [][]byte
}

// LengthField 具备的基础属性
type LengthField struct {
	/*
		大端模式：是指数据的高字节保存在内存的低地址中，而数据的低字节保存在内存的高地址中，地址由小向大增加，而数据从高位往低位放；
		小端模式：是指数据的高字节保存在内存的高地址中，而数据的低字节保存在内存的低地址中，高地址部分权值高，低地址部分权值低，和我们的日常逻辑方法一致。
	*/
	Order               binary.ByteOrder //The byte order: BigEndian or LittleEndian(大小端)
	MaxFrameLength      uint64           //The maximum length of a frame(最大帧长度)
	LengthFieldOffset   int              //The offset of the length field(长度字段偏移量)
	LengthFieldLength   int              //The length of the length field in bytes(长度域字段的字节数)
	LengthAdjustment    int              //The length adjustment(长度调整)
	InitialBytesToStrip int              //The number of bytes to strip from the decoded frame(需要跳过的字节数)
}
