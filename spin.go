package ksync

import (
	"runtime"
	"sync/atomic"
)

type spin int32

func NewSpin() *spin {
	return new(spin)
}

func (s *spin) Lock() {
	for !atomic.CompareAndSwapInt32((*int32)(s), 0, 1) {
		runtime.Gosched()
	}
}

func (s *spin) Unlock() {
	atomic.StoreInt32((*int32)(s), 0)
}

func (s *spin) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(s), 0, 1)
}
