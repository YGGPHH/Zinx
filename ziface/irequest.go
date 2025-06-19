package ziface

// IRequest 为接口, Request 打包客户端请求的连接信息以及请求数据
type IRequest interface {
	GetConnection() IConnection // 获取请求连接信息
	GetData() []byte            // 获取请求消息的数据
	GetMsgID() uint32           // 获取请求的 id
}
