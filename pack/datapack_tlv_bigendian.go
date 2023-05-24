package pack

import (
	"bytes"
	"encoding/binary"
	"github.com/oldbai555/baix/berr"
	"github.com/oldbai555/baix/conf"
	"github.com/oldbai555/baix/iface"
	"github.com/oldbai555/lbtool/log"
)

const defaultHeaderLen uint32 = 8

type dataPack struct{}

// 封包拆包实例初始化方法
func newDataPack() iface.IDataPack {
	return &dataPack{}
}

// GetHeadLen 获取包头长度方法
func (dp *dataPack) GetHeadLen() uint32 {
	//ID uint32(4 bytes) +  DataLen uint32(4 bytes)
	return defaultHeaderLen
}

// Pack 封包方法,压缩数据
func (dp *dataPack) Pack(msg iface.IMessage) ([]byte, error) {
	// 创建一个存放bytes字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	// Write the data length
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetMsgID()); err != nil {
		log.Errorf("err is %v", err)
		return nil, err
	}

	// Write the message ID
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetDataLen()); err != nil {
		log.Errorf("err is %v", err)
		return nil, err
	}

	// Write the data
	if err := binary.Write(dataBuff, binary.BigEndian, msg.GetData()); err != nil {
		log.Errorf("err is %v", err)
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

// Unpack 拆包方法,解压数据
func (dp *dataPack) Unpack(binaryData []byte) (iface.IMessage, error) {
	// Create an ioReader for the input binary data
	dataBuff := bytes.NewReader(binaryData)

	// 只解压head的信息，得到dataLen和msgID
	msg := &Message{}

	// Read the data length
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.ID); err != nil {
		return nil, err
	}

	// Read the message ID
	if err := binary.Read(dataBuff, binary.BigEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 判断dataLen的长度是否超出我们允许的最大包长度
	if conf.GlobalConfig.MaxPacketSize > 0 && msg.GetDataLen() > conf.GlobalConfig.MaxPacketSize {
		return nil, berr.ErrToLargeMsgData
	}

	// 这里只需要把head的数据拆包出来就可以了，然后再通过head的长度，再从conn读取一次数据
	return msg, nil
}
