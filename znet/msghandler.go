package znet

import (
	"fmt"
	"strconv"
	"zinx/settings"
	"zinx/ziface"
)

type MsgHandle struct {
	Apis           map[uint32]ziface.IRouter // map 存放每个 MsgId 对应的处理方法
	WorkerPoolSize uint32                    // 业务工作 worker 池的数量
	TaskQueue      []chan ziface.IRequest    // worker 负责取任务的消息队列
}

var _ ziface.IMsgHandle = (*MsgHandle)(nil)

func NewMsgHandle() *MsgHandle {
	return &MsgHandle{
		Apis:           make(map[uint32]ziface.IRouter),
		WorkerPoolSize: settings.Conf.WorkerPoolSize,
		TaskQueue:      make([]chan ziface.IRequest, settings.Conf.WorkerPoolSize),
	}
}

// 立即以非阻塞的方式处理消息
func (mh *MsgHandle) DoMsgHandler(request ziface.IRequest) {
	handler, ok := mh.Apis[request.GetMsgID()]
	if !ok {
		fmt.Println("api msgId = ", request.GetMsgID(), " is not FOUND!")
		return
	}

	// 执行 Router 的 Handler
	handler.PreHandle(request)
	handler.Handle(request)
	handler.PostHandle(request)
}

// 为某条消息添加具体的处理逻辑
func (mh *MsgHandle) AddRouter(msgId uint32, router ziface.IRouter) {
	// 判断当前 msg 绑定的 API 处理方法是否已经存在
	if _, ok := mh.Apis[msgId]; ok {
		panic("repeated api, msgId = " + strconv.Itoa(int(msgId)))
	}

	// 添加 msg 与 api 的绑定关系
	mh.Apis[msgId] = router
	fmt.Println("Add api msgId = ", msgId)
}

// StartOneWorker 启动一个 Worker 的工作流程
func (mh *MsgHandle) StartOneWorker(workerID int, taskQueue chan ziface.IRequest) {
	fmt.Println("Worker ID = ", workerID, " is started.")
	for {
		select {
		case request := <-taskQueue:
			mh.DoMsgHandler(request)
		}
	}
}

// StartWorkerPool 启动 worker 工作池
func (mh *MsgHandle) StartWorkerPool() {
	// 遍历需要启动的 worker, 依次启动
	for i := 0; i < int(mh.WorkerPoolSize); i++ {
		// 一个 worker 被启动时, 给当前的 worker 对应的任务队列开辟空间
		mh.TaskQueue[i] = make(chan ziface.IRequest, settings.Conf.MaxWorkerTaskLen)
		// 启动当前 worker, 阻塞地等待对应的任务队列是否有消息传来
		go mh.StartOneWorker(i, mh.TaskQueue[i])
	}
}

// SendMsgToTaskQueue 将消息交给 TaskQueue, 由 worker 进行处理
func (mh *MsgHandle) SendMsgToTaskQueue(request ziface.IRequest) {
	// 根据 ConnID 来分配当前的连接应该由哪个 worker 负责处理
	// 使用轮询分配法则

	// 得到需要处理此条连接地 workerID
	workerID := request.GetConnection().GetConnID() % mh.WorkerPoolSize
	fmt.Println("Add ConnID = ", request.GetConnection().GetConnID(), " request msgID = ", request.GetMsgID(), "to workerID = ", workerID)
	// 将请求消息发送给任务队列
	mh.TaskQueue[workerID] <- request
}
