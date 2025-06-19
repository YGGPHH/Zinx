package znet

import (
	"errors"
	"fmt"
	"sync"
	"zinx/ziface"
)

type ConnManager struct {
	connections map[uint32]ziface.IConnection // 管理连接的信息
	connLock    sync.RWMutex                  // 读写连接的读写锁
}

func NewConnManager() *ConnManager {
	return &ConnManager{
		connections: make(map[uint32]ziface.IConnection),
	}
}

// Add 添加连接
func (connMgr *ConnManager) Add(conn ziface.IConnection) {
	// 保护共享资源 Map, 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 将 conn 连接添加到 ConnManager
	connMgr.connections[conn.GetConnID()] = conn

	fmt.Println("connection add to ConnManager successfully: conn num = ", connMgr.Len())
}

// Remove 删除连接
func (connMgr *ConnManager) Remove(conn ziface.IConnection) {
	// 保护共享资源 map, 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 删除连接信息
	delete(connMgr.connections, conn.GetConnID())

	fmt.Println("connection Remove connID = ", conn.GetConnID(), " successfully: conn num = ", connMgr.Len())
}

// Get 利用 ConnID 获取连接
func (connMgr *ConnManager) Get(connID uint32) (ziface.IConnection, error) {
	// 保护共享资源 Map, 加读锁
	connMgr.connLock.RLock()
	defer connMgr.connLock.RUnlock()

	if conn, ok := connMgr.connections[connID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection not found")
	}
}

// Len 获取当前连接个数
func (connMgr *ConnManager) Len() int {
	return len(connMgr.connections)
}

// ClearConn 停止并清除当前所有连接
func (connMgr *ConnManager) ClearConn() {
	// 保护共享资源 Map, 加写锁
	connMgr.connLock.Lock()
	defer connMgr.connLock.Unlock()

	// 停止并删除全部的连接信息
	for connID, conn := range connMgr.connections {
		// 停止
		conn.Stop()
		// 删除
		delete(connMgr.connections, connID)
	}

	fmt.Println("Clear All Connections successfully: conn num = ", connMgr.Len())
}
