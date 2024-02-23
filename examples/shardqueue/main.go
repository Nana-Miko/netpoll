package main

import (
	"context"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nana-miko/netpoll"
	"github.com/nana-miko/netpoll/mux"
	"github.com/stretchr/testify/assert"
)

// 必须使用linux
func main() {
	t := &testingT{}
	//go func() {
	//	fmt.Println(http.ListenAndServe("localhost:6060", nil))
	//}()

	addr := "localhost:9123"
	// 发送的消息条数
	num := 10000 * 1000
	stop := make(chan int)

	go func() {
		// 运行netpoll服务端
		assert.NoError(t, runServe(addr, num, stop))
	}()
	time.Sleep(10 * time.Millisecond)
	// 连接服务端
	connection, err := netpoll.NewDialer().DialConnection("tcp", addr, 5*time.Second)
	assert.NoError(t, err)

	queue := mux.NewShardQueue(409600, connection)
	assert.NoError(t, err)

	meta := NewMetaData()
	meta.payload = make([]byte, 100)

	cb := func(err error) {
		assert.NoError(t, err)
	}
	sTime := time.Now()
	for i := 0; i < num; i++ {
		ctx := context.Background()
		// 异步方法
		queue.AsyncWrite(ctx, meta, cb)
	}
	fmt.Println("所有发送完毕！")
	fmt.Println("用时：", time.Since(sTime).String())
	assert.Equal(t, num, <-stop)
	queue.Close()
	wg := sync.WaitGroup{}
	wg.Add(1)
	queue.AsyncWrite(context.Background(), meta, func(err error) {
		assert.Error(t, err)
		wg.Done()
	})
	wg.Wait()

}

/*
// 以下是mux中的ShardQueue基本使用方法

// 实现mux.PayLoad接口
	type myPayLoad struct {
		data []byte
	}

	func (m *myPayLoad) encodeSize() int {
		// 编码需要的长度（定长编码）
		return 4
	}

	func (m *myPayLoad) encode([]byte) error {
		// 实现编码规则
		return nil
	}

// --- mux.SizedEncodable接口实现 ---

// ZeroCopyEncode writer的大小会根据SizedEncodable.EncodedLen()方法进行分配
// writer是ZeroCopy的，不需要考虑内存分配问题

	func (m *myPayLoad) ZeroCopyEncode(writer netpoll.Writer) error {
		// 向writer申请所需要的空间
		ml, err := writer.Malloc(m.EncodedLen())
		if err != nil {
			return err
		}
		err = m.encode(ml)
		if err != nil {
			return err
		}
		// 调用Flush后才会真正写入到writer
		return writer.Flush()
	}


	func (m *myPayLoad) EncodedLen() int {
		// 返回编码需要的空间
		return len(m.data) + m.encodeSize()
	}

	func quickStart() {
		// 创建的netpoll连接
		var conn netpoll.Connection

		// 新建AsyncWriter，quota大小由业务数据决定，一般在业务数据的10-40倍左右
		writer, err := NewShardQueue(40960, conn)
		if err != nil {
			panic(err)
		}

		payload := &myPayLoad{data: []byte("123456")}
		// 定义CallBack
		cb := func(err error) {
			if err != nil {
				panic(err)
			}
		}
		ctx := context.Background()
		context.WithTimeout(ctx, 5*time.Second)

		writer.AsyncWrite(ctx, payload, cb)

		writer.Close()
	}
*/

type handler struct {
	loop netpoll.EventLoop
	// 计数
	count int32
	// 总数
	num    int32
	reader *MetaReader
}

type testingT struct{}

func (testingT) Errorf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

func (h *handler) handle(ctx context.Context, connection netpoll.Connection) error {

	for {
		if atomic.LoadInt32(&h.count) == h.num {
			h.loop.Shutdown(ctx)
			return nil
		}
		_, err := h.reader.ReadMetaFrom(connection)
		if err != nil {
			return err
		}
		//fmt.Println(string(data))
		atomic.AddInt32(&h.count, 1)
		//fmt.Println(atomic.LoadInt32(&h.count))
	}

}

func runServe(addr string, num int, stop chan int) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	h := handler{
		num:    int32(num),
		reader: NewMetaReader(),
	}

	eventLoop, _ := netpoll.NewEventLoop(h.handle)

	h.loop = eventLoop

	_ = eventLoop.Serve(listener)

	stop <- int(h.count)
	return nil
}
