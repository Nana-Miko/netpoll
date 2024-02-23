package netpoll

import "github.com/nana-miko/netpoll/pool"

type CallBack func(err error)

var asyncLinkBufferPool = pool.NewSyncPool[*AsyncLinkBuffer](func() any {
	lb := new(AsyncLinkBuffer)
	lb.LinkBuffer = new(LinkBuffer)
	return lb
})

// AsyncLinkBuffer A LinkBuffer With CallBack
type AsyncLinkBuffer struct {
	*LinkBuffer

	cb CallBack
}

// NewAsyncLinkBuffer 从对象池中获取
func NewAsyncLinkBuffer(size int, cb CallBack) *AsyncLinkBuffer {
	lb := asyncLinkBufferPool.Get()
	lb.Reuse(size)
	lb.cb = cb
	return lb
}

// GetCallBack 获取CallBack
func (b *AsyncLinkBuffer) GetCallBack() CallBack {
	return b.cb
}

func (b *AsyncLinkBuffer) zero() {
	_ = b.Release()
	b.cb = nil
}

// Recycle 回收AsyncLinkBuffer
func (b *AsyncLinkBuffer) Recycle() {
	b.zero()
	asyncLinkBufferPool.Put(b)
}
