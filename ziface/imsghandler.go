package ziface

type IMsgHandle interface {
	DoMsgHandler(request IRequest)          // 立即以非阻塞的方式处理消息
	AddRouter(msgId uint32, router IRouter) // 为消息添加具体的处理逻辑
	StartWorkerPool()                       // 启动 worker 工作池
	SendMsgToTaskQueue(request IRequest)    // 将消息交给 TaskQueue, 由 worker 进行处理
}
