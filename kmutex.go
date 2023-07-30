package ksync

import (
	"sync"
)

type Kmutex struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*nMutex
}

type nMutex struct {
	lock *sync.Mutex
	n    uint64
}

func DefaultKmutex() *Kmutex {
	return &Kmutex{
		l: new(Spin),
		p: &sync.Pool{
			New: func() any {
				return &nMutex{
					lock: new(sync.Mutex),
				}
			},
		},
		m: make(map[any]*nMutex),
	}
}

func NewKmutex(locker ...sync.Locker) *Kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
	}
	return km
}

func (k *Kmutex) Unlock(key any) {
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

func (k *Kmutex) Lock(key any) {
	k.l.Lock()
	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nMutex)
		k.m[key] = kl
	}
	kl.n++
	k.l.Unlock()

	kl.lock.Lock()
}

func (k *Kmutex) TryLock(key any) (ok bool) {
	k.l.Lock()
	defer k.l.Unlock()
	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nMutex)
		k.m[key] = kl
	}

	ok = kl.lock.TryLock()
	if ok {
		kl.n++
	}
	return
}
