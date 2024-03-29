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
	lock Mutex
	n    uint64
}

type Mutex interface {
	Lock()
	TryLock() bool
	Unlock()
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

type KmutexConf func(*Kmutex)

func WithKmutexLocker(new func() Mutex) KmutexConf {
	return func(k *Kmutex) {
		k.p = &sync.Pool{
			New: func() any {
				return &nMutex{
					lock: new(),
				}
			},
		}
	}
}

func NewKmutex(conf ...KmutexConf) *Kmutex {
	km := DefaultKmutex()
	for _, c := range conf {
		c(km)
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
