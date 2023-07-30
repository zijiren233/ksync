package ksync

import (
	"sync"
)

type Kcond struct {
	l *sync.RWMutex
	p *sync.Pool
	m map[any]*kc
}

type kc struct {
	cond *sync.Cond
	n    uint64
}

func DefaultKcond() *Kcond {
	return &Kcond{
		l: &sync.RWMutex{},
		p: &sync.Pool{
			New: func() any {
				return &kc{
					cond: sync.NewCond(new(sync.Mutex)),
				}
			},
		},
		m: make(map[any]*kc),
	}
}

func NewKcond() *Kcond {
	return DefaultKcond()
}

func (k *Kcond) Wait(key any) {
	k.l.Lock()
	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*kc)
		k.m[key] = kl
	}
	kl.n++
	k.l.Unlock()

	kl.cond.L.Lock()
	kl.cond.Wait()
	kl.cond.L.Unlock()

	k.l.Lock()
	kl.n--
	if kl.n == 0 {
		k.p.Put(kl)
		delete(k.m, key)
	}
	k.l.Unlock()
}

func (k *Kcond) Broadcast(key any) {
	k.l.RLock()
	kl, ok := k.m[key]
	k.l.RUnlock()
	if !ok {
		return
	}

	kl.cond.Broadcast()
}

func (k *Kcond) Signal(key any) {
	k.l.RLock()
	kl, ok := k.m[key]
	k.l.RUnlock()
	if !ok {
		return
	}

	kl.cond.Signal()
}
