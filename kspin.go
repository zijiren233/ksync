package ksync

func NewKspin() *Kmutex {
	return NewKmutex(WithKmutexLocker(func() Mutex {
		return new(Spin)
	}))
}
