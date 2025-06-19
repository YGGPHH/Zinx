package znet

import (
	"fmt"
	"net"
	"time"
	"zinx/settings"
	"zinx/ziface"
)

type Server struct {
	Name       string              // Name 为服务器的名称
	IPVersion  string              // IPVersion: IPv4 or other
	IP         string              // IP: 服务器绑定的 IP 地址
	Port       int                 // Port: 服务器绑定的端口
	msgHandler ziface.IMsgHandle   // 将 Router 替换为 MsgHandler, 绑定 MsgId 与对应的处理方法
	ConnMgr    ziface.IConnManager // 当前 Server 的连接管理器

	onConnStart func(conn ziface.IConnection) // Server 在连接创建时的 Hook 函数
	onConnStop  func(conn ziface.IConnection) // Server 在连接删除时的 Hook 函数
}

// 确保 Server 实现了 ziface.IServer 的所有方法
var _ ziface.IServer = (*Server)(nil)

/* =============== 实现 ziface.IServer 接口当中全部的方法 =============== */

// Start 开启 Server 的网络服务
func (s *Server) Start() {
	fmt.Printf("[START] Server listenner at IP: %s, Port %d, is starting\n", s.IP, s.Port)
	fmt.Printf("[Zinx] Version: %s, MaxConn: %d, MaxPacketSize: %d\n",
		settings.Conf.Version,
		settings.Conf.MaxConn,
		settings.Conf.MaxPacketSize)
	// 开启一个 goroutine 去做服务端的 Listener 业务
	go func() {
		// 0. 启动 worker 工作池机制
		s.msgHandler.StartWorkerPool()

		// 1. 获取一个 TCP 的 Addr
		addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
		if err != nil {
			fmt.Println("resolve tcp addr err: ", err)
			return
		}

		// 2. 监听服务器地址
		listener, err := net.ListenTCP(s.IPVersion, addr)
		if err != nil {
			fmt.Println("listen", s.IPVersion, "err", err)
			return
		}

		// 监听成功
		fmt.Println("start Zinx server  ", s.Name, " succ, now listenning...")

		// TODO: server.go 应该有一个自动生成 ID 的方法, 比如 snowflake
		var cid uint32
		cid = 0

		// 3. 启动 Server 网络连接服务
		for {
			// 3.1 阻塞等待客户端建立连接请求
			conn, err := listener.AcceptTCP()
			if err != nil {
				fmt.Println("Accept err", err)
				continue
			}

			// 3.2 Server.Start() 设置服务器最大连接控制, 如果超过最大连接, 则关闭此新的连接
			if s.ConnMgr.Len() >= settings.Conf.MaxConn {
				// 是否可以制定一个类似于 LRUCache 的连接规则 ?
				conn.Close()
				continue
			}

			// 3.3 处理该连接请求的业务方法, 此时应该有 handler 和 conn 是绑定的
			dealConn := NewConnection(s, conn, cid, s.msgHandler)
			cid++

			go dealConn.Start()
		}
	}()
}

func (s *Server) Stop() {
	fmt.Println("[STOP] Zinx server , name ", s.Name)

	// Server.Stop() 将其它需要清理的连接信息或其他信息一并停止或清理
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	s.Start()

	// TODO Server.Serve() 是否在启动服务的时候, 还需要做其它事情呢?

	// 阻塞, 否则 main goroutine 退出, listenner 也将会随之退出
	for {
		time.Sleep(10 * time.Second)
	}
}

func (s *Server) AddRouter(msgId uint32, router ziface.IRouter) {
	s.msgHandler.AddRouter(msgId, router)
	fmt.Println("Add Router succ! msgId = ", msgId)
}

func (s *Server) GetConnMgr() ziface.IConnManager {
	return s.ConnMgr
}

// NewServer 将创建一个服务器的 Handler
func NewServer() ziface.IServer {
	s := &Server{
		Name:       settings.Conf.Name,
		IPVersion:  "tcp4",
		IP:         settings.Conf.Host,
		Port:       settings.Conf.Port,
		msgHandler: NewMsgHandle(),
		ConnMgr:    NewConnManager(),
	}

	return s
}

func (s *Server) SetOnConnStart(hookFunc func(ziface.IConnection)) {
	s.onConnStart = hookFunc
}

func (s *Server) SetOnConnStop(hookFunc func(ziface.IConnection)) {
	s.onConnStop = hookFunc
}

func (s *Server) CallOnConnStart(conn ziface.IConnection) {
	if s.onConnStart != nil {
		fmt.Println("---> CallOnConnStart ...")
		s.onConnStart(conn)
	}
}

func (s *Server) CallOnConnStop(conn ziface.IConnection) {
	if s.onConnStop != nil {
		fmt.Println("---> CallOnConnStop ...")
		s.onConnStop(conn)
	}
}
