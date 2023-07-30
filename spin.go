package ksync

import (
	"runtime"
	"sync/atomic"
)

type Spin int32

func NewSpin() *Spin {
	return new(Spin)
}

func (s *Spin) Lock() {
	for !s.TryLock() {
		runtime.Gosched()
	}
}

func (s *Spin) Unlock() {
	atomic.StoreInt32((*int32)(s), 0)
}

func (s *Spin) TryLock() bool {
	return atomic.CompareAndSwapInt32((*int32)(s), 0, 1)
}
