package thread

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type SpinLock uint32

func (sl *SpinLock) Lock() {
	for !atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1) {
		runtime.Gosched()
	}
}

func (sl *SpinLock) Unlock() {
	atomic.StoreUint32((*uint32)(sl), 0)
}

// TryLock
// return:
// - true: lock successfully
// - false: lock failed
func (sl *SpinLock) TryLock() bool {
	return atomic.CompareAndSwapUint32((*uint32)(sl), 0, 1)
}

// NewSpinLock creates a spin lock,
// which is not re-entrant.
func NewSpinLock() sync.Locker {
	var lock SpinLock
	return &lock
}
