package netpoll

import (
	"github.com/bytedance/gopkg/util/gopool"
)

// goPool的默认协程池是有Cap上限的，这里new一个无上限的池
var goroutinePool = gopool.NewPool("netpoll.goPool", -1, gopool.NewConfig())

// Go 在协程池中运行Func
func Go(f func()) {
	goroutinePool.Go(f)
}
