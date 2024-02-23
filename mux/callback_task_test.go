package mux

import (
	"errors"
	"fmt"
	"github.com/nana-miko/netpoll"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCbTaskList_Foreach(t *testing.T) {
	count := 0
	num := 100
	tTerr := errors.New("err")
	cb := func(err error) {
		assert.Error(t, err)
		count++
	}
	cbList := NewCbTaskList()
	cbList2 := NewCbTaskList()
	for i := 0; i < num; i++ {
		cbList.Append(cb)
		cbList2.Append(cb)
	}
	cbList.Foreach(tTerr)
	cbList2.Foreach(tTerr)

	assert.Equal(t, num*2, count)

}

// 测试链表的GC是否正常
func TestCbTaskList_GC(t *testing.T) {
	kb15 := uint64(1024 * 15)
	stop := make(chan struct{})
	var startMem, endMem runtime.MemStats
	cb := func(err error) {
		assert.Error(t, err)
	}

	err := fmt.Errorf("err")
	runtime.GC()
	runtime.ReadMemStats(&startMem)
	netpoll.Go(func() {
		for {
			select {
			case <-stop:
				return
			default:
			}
			cbList := NewCbTaskList()
			for i := 0; i < 10; i++ {
				cbList.Append(cb)
			}
			cbList.Foreach(err)
			cbList.Recycle()
		}
	})
	<-time.After(1 * time.Second)
	runtime.ReadMemStats(&endMem)
	alloc := endMem.TotalAlloc - startMem.TotalAlloc
	t.Logf("链表申请内存：%d %s", alloc, "byte")

	assert.Equal(t, alloc <= kb15, true)

	stop <- struct{}{}

}
