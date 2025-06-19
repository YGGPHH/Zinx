package znet

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"zinx/settings"
	"zinx/ziface"
)

type Connection struct {
	TCPServer    ziface.IServer    // 标记当前 Conn 属于哪个 Server
	Conn         *net.TCPConn      // 当前连接的 socket TCP 套接字
	ConnID       uint32            // 当前连接的 ID, 也可称为 SessionID, 全局唯一
	isClosed     bool              // 当前连接的开启/关闭状态
	Msghandler   ziface.IMsgHandle // 将 Router 替换为消息管理模块
	ExitBuffChan chan bool         // 告知该连接一经退出/停止的 channel
	msgChan      chan []byte       // 无缓冲 channel, 用于读/写两个 goroutine 之间的消息通信
	msgBuffChan  chan []byte       // 定义 msgBuffChan

	property     map[string]interface{} // 连接属性
	propertyLock sync.RWMutex           // 保护连接属性修改的锁
}

// 确保 Connection 实现 ziface.IConenction 方法
var _ ziface.IConnection = (*Connection)(nil)

// NewConnection 创建新的连接
func NewConnection(server ziface.IServer, conn *net.TCPConn, connID uint32, msgHandler ziface.IMsgHandle) *Connection {
	c := &Connection{
		TCPServer:    server,
		Conn:         conn,
		ConnID:       connID,
		isClosed:     false,
		Msghandler:   msgHandler,
		ExitBuffChan: make(chan bool, 1),
		msgChan:      make(chan []byte), // msgChan 初始化
		msgBuffChan:  make(chan []byte, settings.Conf.MaxMsgChanLen),
		property:     make(map[string]interface{}),
	}

	// 将新创建的 Conn 添加到连接管理器中
	c.TCPServer.GetConnMgr().Add(c)

	return c
}

// SetProperty 用于设置连接属性
func (c *Connection) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

// GetProperty 获取连接属性
func (c *Connection) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("No property found")
	}
}

// RemoveProperty 移除连接属性
func (c *Connection) RemoveProperty(key string) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	delete(c.property, key)
}

func (c *Connection) StartWriter() {
	fmt.Println("[Writer Goroutine is running]")
	defer fmt.Println(c.RemoteAddr().String(), "[conn Writer exit!]")

	for {
		select {
		case data := <-c.msgChan:
			if _, err := c.Conn.Write(data); err != nil {
				fmt.Println("Send Data error:", err, " Conn Writer exit~")
				return
			}
		case data, ok := <-c.msgBuffChan:
			if ok {
				if _, err := c.Conn.Write(data); err != nil {
					fmt.Println("Send Buff Data error:", err, " Conn Writer exit")
					return
				}
			} else {
				fmt.Println("msgBuffChan is Closed")
				return
			}
		case <-c.ExitBuffChan:
			// conn 关闭
			return
		}
	}
}

// StartReader 开启处理 conn 读数据的 goroutine
func (c *Connection) StartReader() {
	fmt.Println("Reader Goroutine is running")
	defer fmt.Println(c.RemoteAddr().String(), " conn reader exit !")
	defer c.Stop()

	for {
		// 创建封包拆包的对象
		dp := NewDataPack()

		// 读取客户端的 msg head
		headData := make([]byte, dp.GetHeadLen()) // 注意 GetHeadLen() 返回常量 8, 因为包的头部长度固定
		if _, err := io.ReadFull(c.GetTCPConnection(), headData); err != nil {
			fmt.Println("read msg head error", err)
			c.ExitBuffChan <- true
			return
		}

		// 拆包, 得到 msgid 和 datalen, 并放在 msg 中
		msg, err := dp.Unpack(headData)
		if err != nil {
			fmt.Println("unpack error", err)
			c.ExitBuffChan <- true
			return
		}

		// 根据 dataLen 读取 data, 放在 msg.Data 中
		var data []byte
		if msg.GetDataLen() > 0 {
			data = make([]byte, msg.GetDataLen())
			if _, err := io.ReadFull(c.GetTCPConnection(), data); err != nil {
				fmt.Println("read msg data error", err)
				c.ExitBuffChan <- true
				return
			}
		}
		msg.SetData(data)

		// 得到当前客户端请求的 Request 数据
		req := Request{
			conn: c,
			msg:  msg,
		}

		if settings.Conf.WorkerPoolSize > 0 {
			// 已经启动工作池机制, 将消息交给 Worker 处理
			c.Msghandler.SendMsgToTaskQueue(&req)
		} else {
			// 从绑定好的消息和对应的处理方法中执行 Handle 方法
			go c.Msghandler.DoMsgHandler(&req)
		}
	}
}

// Start 实现 IConnection 中的方法, 它启动连接并让当前连接开始工作
func (c *Connection) Start() {
	// 开启处理该连接读取到客户端数据之后的业务请求
	go c.StartWriter()
	go c.StartReader()

	c.TCPServer.CallOnConnStart(c)

	for {
		select {
		case <-c.ExitBuffChan:
			// 得到退出消息则不再阻塞
			return
		}
	}
}

// Stop 停止连接, 结束当前连接状态
func (c *Connection) Stop() {
	fmt.Println("Conn Stop()... ConnID = ", c.ConnID)
	// 1. 如果当前连接已经关闭
	if c.isClosed == true {
		return
	}
	c.isClosed = true

	// Connection Stop() 如果用户注册了该连接的关闭回调业务, 那么应该在此刻显式调用
	c.TCPServer.CallOnConnStop(c)

	// 关闭 socket 连接
	c.Conn.Close()
	// 通知从缓冲队列读数据的业务, 该链接已经关闭
	c.ExitBuffChan <- true

	// 将连接从管理器中删除
	c.TCPServer.GetConnMgr().Remove(c)

	// 关闭该链接全部管道
	close(c.ExitBuffChan)
	close(c.msgBuffChan)
}

// GetTCPConnection 从当前连接获取原始的 socket TCPConn
func (c *Connection) GetTCPConnection() *net.TCPConn {
	return c.Conn
}

// GetConnID 获取当前连接的 ID
func (c *Connection) GetConnID() uint32 {
	return c.ConnID
}

// RemoteAddr 获取远程客户端的地址信息
func (c *Connection) RemoteAddr() net.Addr {
	return c.Conn.RemoteAddr()
}

func (c *Connection) SendMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send msg")
	}

	// 将 data 封包
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg ")
	}

	// 将 data 发送
	c.msgChan <- msg
	return nil
}

func (c *Connection) SendBuffMsg(msgId uint32, data []byte) error {
	if c.isClosed == true {
		return errors.New("Connection closed when send buff msg")
	}

	// 将 data 封包并发送
	dp := NewDataPack()
	msg, err := dp.Pack(NewMsgPackage(msgId, data))
	if err != nil {
		fmt.Println("Pack error msg id = ", msgId)
		return errors.New("Pack error msg")
	}

	c.msgBuffChan <- msg

	return nil
}
