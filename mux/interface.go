package mux

import (
	"context"
	"io"

	"github.com/nana-miko/netpoll"
)

// AsyncWriter 异步Writer
// 真正的连接写入由Writer内部管理
type AsyncWriter interface {
	// AsyncWrite 异步写方法
	AsyncWrite(ctx context.Context, data SizedEncodable, cb netpoll.CallBack)
	io.Closer
}

// SizedEncodable 是一个接口，定义了对数据进行编码的方法
type SizedEncodable interface {
	// ZeroCopyEncode 对其进行编码（线程安全）
	ZeroCopyEncode(writer netpoll.Writer) error
	// EncodedLen 编码所需要的空间
	EncodedLen() int
}
