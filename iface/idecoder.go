package iface

// 解码器

type IDecoder interface {
	IInterceptor
	GetLengthField() *LengthField
}
