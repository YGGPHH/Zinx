package ziface

type IRouter interface {
	PreHandle(request IRequest)  // 处理 conn 业务之前的钩子方法
	Handle(request IRequest)     // 处理 conn 业务的方法
	PostHandle(request IRequest) // 处理 conn 业务之后的钩子方法
}
