package pool

import (
	"github.com/tinboxw/sword/thread"
	"time"
)

const (
	Size = 2
)

type Pool[T any] struct {
	index      uint
	version    uint
	changed    bool
	ori        *T
	buckets    [Size]*T
	updateTime time.Time
}

func New[T any](src [Size]*T) *Pool[T] {
	return &Pool[T]{
		index:      0,
		version:    0,
		changed:    false,
		buckets:    src,
		updateTime: time.Now(),
	}
}

func (x *Pool[T]) Fetch() *T {
	if x.changed {
		x.changed = false
		x.ori = x.Get()
	}
	return x.ori
}

func (x *Pool[T]) Get() *T       { return x.buckets[x.getIndex()] }
func (x *Pool[T]) GetNext() *T   { return x.buckets[x.getNextIndex()] }
func (x *Pool[T]) Version() uint { return x.version }

func (x *Pool[T]) Update(src *T) *T {
	next := x.getNextIndex()
	old := x.buckets[next]
	x.buckets[next] = src
	x.TakeNext()
	return old
}

func (x *Pool[T]) TakeNext() {
	x.version++
	x.index = x.getNextIndex()
	x.updateTime = x.updateTime.Add(time.Since(x.updateTime))
	x.changed = true
}

func (x *Pool[T]) getIndex() uint     { return x.index }
func (x *Pool[T]) getNextIndex() uint { return (x.index + 1) % Size }

type TsPool[T any] struct {
	*Pool[T]
	lock thread.SpinLock
}

func NewTsPool[T any](src [2]*T) *TsPool[T] {
	return &TsPool[T]{
		Pool: New(src),
	}
}

type Element[T any] struct {
	value *T
	pool  *TsPool[T]
}

func (e Element[T]) Read() *T  { return e.value }
func (e Element[T]) ReadDone() { e.pool.ReadDone() }

func (x *TsPool[T]) Read() *Element[T] {
	return &Element[T]{
		value: x.ReadDo(),
		pool:  x,
	}
}

func (x *TsPool[T]) ReadDo() *T       { x.lock.Lock(); return x.Get() }
func (x *TsPool[T]) ReadDone()        { x.lock.Unlock() }
func (x *TsPool[T]) GetWriteable() *T { return x.GetNext() }
func (x *TsPool[T]) TryFlush(fnSuccess func(cur, next *T)) bool {
	if x.lock.TryLock() {
		x.TakeNext()
		cur := x.Get()
		next := x.GetNext()
		if fnSuccess != nil {
			fnSuccess(cur, next)
		}
		x.lock.Unlock()
		return true
	}

	return false
}
