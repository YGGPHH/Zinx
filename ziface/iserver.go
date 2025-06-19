package ziface

// 定义服务器接口
type IServer interface {
	Start()                                 // Start 启动服务器方法
	Stop()                                  // Stop 停止服务器方法
	Serve()                                 // Serve 开启服务器方法
	AddRouter(msgId uint32, router IRouter) // 路由功能: 给当前服务注册一个路由业务方法
	GetConnMgr() IConnManager               // 得到连接管理器

	SetOnConnStart(func(IConnection)) // 设置该 Server 在连接创建时的 hook 函数
	SetOnConnStop(func(IConnection))  // 设置该 Server 在连接断开时的 hook 函数
	CallOnConnStart(conn IConnection) // 调用连接 onConnStart Hook 函数
	CallOnConnStop(conn IConnection)  // 调用连接 onConnStop Hook 函数
}
