package ksync

import (
	"sync"
)

type Kspin struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*nSpinLock
}

type nSpinLock struct {
	lock *Spin
	n    uint64
}

func DefaultKspin() *Kspin {
	return &Kspin{
		l: new(Spin),
		p: &sync.Pool{
			New: func() any {
				return &nSpinLock{
					lock: new(Spin),
				}
			},
		},
		m: make(map[any]*nSpinLock),
	}
}

func NewKspin(locker ...sync.Locker) *Kspin {
	kl := DefaultKspin()
	for _, lock := range locker {
		kl.l = lock
	}
	return kl
}

func (k *Kspin) Unlock(key any) {
	k.l.Lock()
	defer k.l.Unlock()

	kl, ok := k.m[key]
	if !ok {
		return
	}

	kl.lock.Unlock()

	kl.n--
	if kl.n == 0 {
		k.p.Put(kl)
		delete(k.m, key)
	}
}

func (k *Kspin) Lock(key any) {
	k.l.Lock()
	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nSpinLock)
		k.m[key] = kl
	}
	kl.n++
	k.l.Unlock()

	kl.lock.Lock()
}

func (k *Kspin) TryLock(key any) (ok bool) {
	k.l.Lock()
	defer k.l.Unlock()
	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nSpinLock)
		k.m[key] = kl
	}

	ok = kl.lock.TryLock()
	if ok {
		kl.n++
	}
	return
}
