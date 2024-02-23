package netpoll

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/panjf2000/ants/v2"
)

var runt = 3 * time.Second

// 12412275
func TestAntsWithFunc(t *testing.T) {
	stop := make(chan struct{})
	p, _ := ants.NewPoolWithFunc(-1, testF)
	num := int64(0)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			_ = p.Invoke(num)
		}
	}()
	<-time.After(runt)
	stop <- struct{}{}
	t.Log("Ants Func:", count)
}

// 10862209
func TestAnts(t *testing.T) {
	p, _ := ants.NewPool(-1)
	stop := make(chan struct{})
	num := int64(0)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			p.Submit(func() {
				atomic.AddInt64(&num, 1)
			})
		}
	}()
	<-time.After(runt)
	stop <- struct{}{}
	t.Log("Ants:", num)
}

// 30633608
func TestAntsGoPool(t *testing.T) {
	p := gopool.NewPool("test.pool", -1, gopool.NewConfig())
	stop := make(chan struct{})
	num := int64(0)
	go func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			p.Go(func() {
				atomic.AddInt64(&num, 1)
			})
		}
	}()
	<-time.After(runt)
	stop <- struct{}{}
	t.Log("GoPool:", num)
}

var count int64

func testF(value any) {
	atomic.AddInt64(&count, 1)
}
