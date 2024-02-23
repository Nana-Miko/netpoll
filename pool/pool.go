package pool

import "sync"

type Pool[T any] interface {
	Get() T
	Put(T)
}

type tPool[T any] struct {
	pool sync.Pool
}

func (p *tPool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *tPool[T]) Put(t T) {
	p.pool.Put(t)
}

func NewSyncPool[T any](f func() any) Pool[T] {
	return &tPool[T]{
		pool: sync.Pool{New: f},
	}
}
