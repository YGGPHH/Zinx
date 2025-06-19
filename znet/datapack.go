package znet

import (
	"bytes"
	"encoding/binary"
	"errors"
	"zinx/settings"
	"zinx/ziface"
)

// DataPack 为用于封包和拆包的类, 暂时不需要成员
type DataPack struct{}

var _ ziface.IDataPack = (*DataPack)(nil)

// NewDataPack 封包拆包实例的初始化方法
func NewDataPack() *DataPack {
	return &DataPack{}
}

// GetHeadLen 获取包头长度
func (dp *DataPack) GetHeadLen() uint32 {
	// Id uint32(4 bytes) + DataLen uint32(4 bytes)
	return 8
}

// Pack 为封包方法
func (dp *DataPack) Pack(msg ziface.IMessage) ([]byte, error) {
	// 创建一个存放 byte 字节的缓冲
	dataBuff := bytes.NewBuffer([]byte{})

	// 写 dataLen
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetDataLen()); err != nil {
		return nil, err
	}

	// 写 msgID
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetMsgId()); err != nil {
		return nil, err
	}

	// 写 data 数据
	if err := binary.Write(dataBuff, binary.LittleEndian, msg.GetData()); err != nil {
		return nil, err
	}

	return dataBuff.Bytes(), nil
}

func (dp *DataPack) Unpack(binaryData []byte) (ziface.IMessage, error) {
	// 创建一个用于输入二进制数据的 ioReader
	dataBuff := bytes.NewReader(binaryData)

	// 只解压 head 信息, 得到 dataLen 和 msgID
	msg := &Message{}

	// 读 dataLen
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.DataLen); err != nil {
		return nil, err
	}

	// 读 msgID
	if err := binary.Read(dataBuff, binary.LittleEndian, &msg.Id); err != nil {
		return nil, err
	}

	// 判断 dataLen 的长度是否超过了我们允许的最大包长度
	if settings.Conf.MaxPacketSize > 0 && msg.DataLen > settings.Conf.MaxPacketSize {
		return nil, errors.New("Too large msg data received")
	}

	// 此处只需要把 head 的数据拆包出来即可, 再通过 head 的长度, 从 conn 读取一次数据
	return msg, nil
}
