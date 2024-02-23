package main

import (
	"encoding/binary"
	"errors"
	"github.com/nana-miko/netpoll"
	"github.com/nana-miko/netpoll/mux"
	"io"
	"math"
)

const (
	sizeOfLength = 4
)

var ErrMarshal = errors.New("data size exceeds the limit of uint32")

// MetaData 使用长度+数据的编码
type MetaData struct {
	payload []byte
}

// 实现SizedEncodable接口

func (m *MetaData) ZeroCopyEncode(wr netpoll.Writer) error {
	dataLen := len(m.payload)
	if dataLen > math.MaxUint32 {
		return ErrMarshal
	}
	if dataLen <= 0 {
		return mux.ErrDataEmpty
	}

	// 预留4字节用于存储长度信息
	mc, err := wr.Malloc(sizeOfLength)
	if err != nil {
		return err
	}
	// 将长度信息写入前4个字节
	binary.LittleEndian.PutUint32(mc, uint32(dataLen))
	_, err = wr.WriteBinary(m.payload)
	if err != nil {
		return err
	}

	return wr.Flush()
}

func (m *MetaData) EncodedLen() int {
	return len(m.payload) + sizeOfLength
}

func NewMetaData() *MetaData {
	return new(MetaData)
}

// MetaReader 用于从字节流中读取由长度(4 byte)+数据组成的元数据
type MetaReader struct {
	lenSli  *SmartSlice[byte]
	dataSli *SmartSlice[byte]
}

func NewMetaReader() *MetaReader {
	md := MetaReader{
		lenSli:  &SmartSlice[byte]{},
		dataSli: &SmartSlice[byte]{},
	}
	md.lenSli.Make(sizeOfLength)
	return &md
}

// ReadMetaFrom 从Reader中读取一个元数据
// 返回的[]byte是对MetaReader内部数据结构的引用对象，有可能在下次读取的时候被修改
// 因此在Read的事件循环内,需要保证在下次循环开始前该[]byte被消费完毕
func (m *MetaReader) ReadMetaFrom(r io.Reader) ([]byte, error) {
	m.lenSli.Clear()
	m.dataSli.Clear()
	for m.lenSli.Available() > 0 {
		read, err := r.Read(m.lenSli.Malloc(m.lenSli.Available()))
		if err != nil {
			return nil, err
		}
		m.lenSli.Flush(read)
	}
	length := binary.LittleEndian.Uint32(m.lenSli.Slice())
	m.dataSli.Make(int(length))
	for m.dataSli.Available() > 0 {
		read, err := r.Read(m.dataSli.Malloc(m.dataSli.Available()))
		if err != nil {
			return m.dataSli.Slice(), err
		}
		m.dataSli.Flush(read)
	}
	return m.dataSli.Slice(), nil
}
