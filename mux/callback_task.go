package mux

import (
	"github.com/nana-miko/netpoll"
	"github.com/nana-miko/netpoll/pool"
)

var cbTaskPool pool.Pool[*cbTaskNode]
var cbTaskListPool pool.Pool[*CbTaskList]

func init() {
	cbTaskPool = pool.NewSyncPool[*cbTaskNode](func() any {
		return new(cbTaskNode)
	})
	cbTaskListPool = pool.NewSyncPool[*CbTaskList](func() any {
		return new(CbTaskList)
	})
}

// cb链表的节点
type cbTaskNode struct {
	cb netpoll.CallBack

	next *cbTaskNode
}

// 清空节点
func (t *cbTaskNode) zero() {
	t.cb = nil
	t.next = nil
}

// Recycle 回收节点
func (t *cbTaskNode) Recycle() {
	t.zero()
	cbTaskPool.Put(t)
}

// CbTaskList cb链表
// 用于执行一批的cb
type CbTaskList struct {
	head *cbTaskNode
	tail *cbTaskNode
}

// NewCbTaskList 从对象池中获取
func NewCbTaskList() *CbTaskList {
	return cbTaskListPool.Get()
}

// Append 向链表添加cb
func (l *CbTaskList) Append(cb netpoll.CallBack) {
	node := cbTaskPool.Get()
	node.cb = cb
	if l.head == nil {
		l.head = node
		l.tail = node
	} else {
		l.tail.next = node
		l.tail = node
	}
}

// 清空链表
func (l *CbTaskList) zero() {
	l.head = nil
	l.tail = nil

}

// Recycle 回收链表
func (l *CbTaskList) Recycle() {
	l.zero()
	cbTaskListPool.Put(l)
}

// Foreach 遍历执行cb
func (l *CbTaskList) Foreach(err error) {
	if l.head == nil {
		return
	}
	current := l.head
	// 前驱节点
	var prev *cbTaskNode
	for {
		if prev != nil {
			// 回收前驱节点
			prev.Recycle()
		}
		if current.cb != nil {
			// 调用回调
			current.cb(err)
		}
		prev = current
		if current.next == nil {
			// 表示当前节点已是尾节点
			// 回收当前节点，遍历结束
			current.Recycle()
			break
		}
		// 指向后驱节点
		current = current.next
	}
}
