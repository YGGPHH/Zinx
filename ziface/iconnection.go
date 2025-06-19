package ziface

import "net"

type IConnection interface {
	Start()                                      // 启动连接
	Stop()                                       // 停止连接
	GetConnID() uint32                           // 获取远程客户端地址信息
	GetTCPConnection() *net.TCPConn              // 从当前连接获取原始的 socket TCPConn
	RemoteAddr() net.Addr                        // 获取远程客户端地址信息
	SendMsg(msgId uint32, data []byte) error     // 直接将 Message 数据发给远程的 TCP 客户端
	SendBuffMsg(msgId uint32, data []byte) error // 添加带缓冲的发送消息接口

	SetProperty(key string, value interface{})   // 设置连接属性
	GetProperty(key string) (interface{}, error) // 获取连接属性
	RemoveProperty(key string)                   // 移除连接属性
}

// HandFunc 定义了一个统一处理连接业务的接口
type HandFunc func(*net.TCPConn, []byte, int) error
