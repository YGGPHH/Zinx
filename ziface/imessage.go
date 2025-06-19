package ziface

type IMessage interface {
	GetDataLen() uint32 // 获取消息数据段的长度
	GetMsgId() uint32   // 获取消息 ID
	GetData() []byte    // 获取消息内容

	SetMsgId(uint32)   // 设置消息 ID
	SetData([]byte)    // 设置消息内容
	SetDataLen(uint32) // 设置消息数据段的长度
}
