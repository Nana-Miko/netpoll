package main

// SmartSlice 智能切片，拥有自动扩容，空间重用等特性
// 线程不安全
type SmartSlice[T any] struct {
	// 底层切片
	sl []T
	// 数据长度
	length int
	// 切片容量
	cap int
}

// Make 初始化切片
// TODO 当某个Cap过大时，会使总体的Slice容量过大，无法得到释放，缺乏一种缩容机制，目前可通过主动调用TrimToSize进行缩容
func (d *SmartSlice[T]) Make(cap int) {
	d.make(cap)
	d.Clear()
}

func (d *SmartSlice[T]) make(cap int) {
	if d.sl == nil {
		d.sl = make([]T, cap)
	}
	if len(d.sl) < cap {
		d.sl = append(d.sl, make([]T, cap-len(d.sl))...)
	}
	d.cap = cap
}

// Clear 清空切片
func (d *SmartSlice[T]) Clear() {
	d.length = 0
}

// Len 切片有效数据长度
func (d *SmartSlice[T]) Len() int {
	return d.length
}

// Cap 切片容量
func (d *SmartSlice[T]) Cap() int {
	return d.cap
}

// RealCap 底层切片容量（实际占用的大小）
func (d *SmartSlice[T]) RealCap() int {
	return cap(d.sl)
}

// AppendFrom 在切片末尾追加
func (d *SmartSlice[T]) AppendFrom(db *SmartSlice[T]) {
	temp := d.Malloc(db.length)
	copy(temp, db.Slice())
	d.Flush(db.length)
}

// CopyFrom 在切片末尾追加
func (d *SmartSlice[T]) CopyFrom(sl []T) {
	temp := d.Malloc(len(sl))
	copy(temp, sl)
	d.Flush(len(sl))
}

// Slice 获取切片内容
func (d *SmartSlice[T]) Slice() []T {
	return d.sl[:d.length]
}

// SliceBetween 获取指定切片内容
func (d *SmartSlice[T]) SliceBetween(start int, end int) []T {
	return d.sl[start:end]
}

// SliceCopy 获取切片内容返回一个新的切片
func (d *SmartSlice[T]) SliceCopy() []T {
	temp := make([]T, d.length)
	copy(temp, d.sl[:d.length])
	return temp
}

// SliceTo 将切片内容copy到另一个切片中
func (d *SmartSlice[T]) SliceTo(sl []T) {
	copy(sl, d.sl[:d.length])
}

func (d *SmartSlice[T]) IsFull() bool {
	if d.cap == d.length {
		return true
	}
	return false
}

func (d *SmartSlice[T]) IsEmpty() bool {
	if d.length == 0 {
		return true
	}
	return false
}

// Malloc 从SmartSlice申请一个新的连续内存空间
// 只有在调用Flush方法后，SmartSlice的长度才会被更新
func (d *SmartSlice[T]) Malloc(size int) []T {
	if size > d.Available() {
		d.make(d.cap + size - d.Available())
	}
	return d.sl[d.length : d.length+size]
}

func (d *SmartSlice[T]) Flush(size int) {
	d.length += size
}

// Available 切片剩余空间长度（当Malloc的size大于该值时就回触发扩容机制）
func (d *SmartSlice[T]) Available() int {
	return d.cap - d.length
}

// TrimToSize 将切片的容量缩小到与当前长度匹配
func (d *SmartSlice[T]) TrimToSize() {
	d.sl = d.SliceCopy()
	d.cap = d.length
}
