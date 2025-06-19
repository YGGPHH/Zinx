package znet

import "zinx/ziface"

type Message struct {
	Id      uint32 // 消息的 ID
	DataLen uint32 // 消息的长度
	Data    []byte // 消息的内容
}

var _ ziface.IMessage = (*Message)(nil)

// NewMsgPackage 创建一个 Message 消息
func NewMsgPackage(id uint32, data []byte) *Message {
	return &Message{
		Id:      id,
		DataLen: uint32(len(data)),
		Data:    data,
	}
}

// GetDataLen 获取消息数据段的长度
func (msg *Message) GetDataLen() uint32 {
	return msg.DataLen
}

// GetMsgId 获取消息 Id
func (msg *Message) GetMsgId() uint32 {
	return msg.Id
}

// GetData 获取消息内容
func (msg *Message) GetData() []byte {
	return msg.Data
}

// SetDataLen 设置数据段长度
func (msg *Message) SetDataLen(len uint32) {
	msg.DataLen = len
}

// SetMsgId 设置消息 Id
func (msg *Message) SetMsgId(msgId uint32) {
	msg.Id = msgId
}

// SetData 设置消息内容
func (msg *Message) SetData(data []byte) {
	msg.Data = data
}
