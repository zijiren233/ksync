package ksync

import (
	"sync"
)

type Kmutex struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*klock
}

type klock struct {
	lock *sync.Mutex
	n    uint64
}

func DefaultKmutex() *Kmutex {
	return &Kmutex{
		l: &sync.Mutex{},
		p: &sync.Pool{
			New: func() any {
				return &klock{
					lock: &sync.Mutex{},
				}
			},
		},
		m: make(map[any]*klock),
	}
}

func NewKmutex(locker ...sync.Locker) *Kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
	}
	return km
}

func (km *Kmutex) Unlock(key any) {
	km.l.Lock()
	defer km.l.Unlock()

	kl, ok := km.m[key]
	if !ok {
		return
	}

	kl.n--
	if kl.n == 0 {
		km.p.Put(kl)
		delete(km.m, key)
	}

	kl.lock.Unlock()
}

func (km *Kmutex) Lock(key any) {
	km.l.Lock()
	kl, ok := km.m[key]
	if !ok {
		kl = km.p.Get().(*klock)
		km.m[key] = kl
	}
	kl.n++
	km.l.Unlock()

	kl.lock.Lock()
}

func (km *Kmutex) TryLock(key any) (ok bool) {
	km.l.Lock()
	defer km.l.Unlock()
	kl, ok := km.m[key]
	if !ok {
		kl = km.p.Get().(*klock)
		km.m[key] = kl
	}

	ok = kl.lock.TryLock()
	if ok {
		kl.n++
	}
	return
}
