package ksync

import (
	"sync"
)

type Krwmutex struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*nRWMutex
}

type nRWMutex struct {
	lock RWMutex
	n    uint64
}

type RWMutex interface {
	Lock()
	RLock()
	RLocker() sync.Locker
	RUnlock()
	TryLock() bool
	TryRLock() bool
	Unlock()
}

func DefaultKrwmutex() *Krwmutex {
	return &Krwmutex{
		l: new(Spin),
		p: &sync.Pool{
			New: func() any {
				return &nRWMutex{
					lock: new(sync.RWMutex),
				}
			},
		},
		m: make(map[any]*nRWMutex),
	}
}

type KrwmutexConf func(*Krwmutex)

func WithKrwmutexLocker(new func() RWMutex) KrwmutexConf {
	return func(k *Krwmutex) {
		k.p = &sync.Pool{
			New: func() any {
				return &nRWMutex{
					lock: new(),
				}
			},
		}
	}
}

func NewKrwmutex(conf ...KrwmutexConf) *Krwmutex {
	km := DefaultKrwmutex()
	for _, c := range conf {
		c(km)
	}
	return km
}

func (k *Krwmutex) Lock(key any) {
	k.l.Lock()

	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nRWMutex)
		k.m[key] = kl
	}
	kl.n++
	k.l.Unlock()

	kl.lock.Lock()
}

func (k *Krwmutex) RLock(key any) {
	k.l.Lock()

	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nRWMutex)
		k.m[key] = kl
	}
	kl.n++
	k.l.Unlock()

	kl.lock.RLock()
}

func (k *Krwmutex) RUnlock(key any) {
	k.l.Lock()
	defer k.l.Unlock()

	kl, ok := k.m[key]
	if !ok {
		return
	}

	kl.lock.RUnlock()

	kl.n--
	if kl.n == 0 {
		k.p.Put(kl)
		delete(k.m, key)
	}
}

func (k *Krwmutex) TryLock(key any) (ok bool) {
	k.l.Lock()
	defer k.l.Unlock()

	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nRWMutex)
		k.m[key] = kl
	}

	ok = kl.lock.TryLock()
	if ok {
		kl.n++
	}
	return
}

func (k *Krwmutex) TryRLock(key any) (ok bool) {
	k.l.Lock()
	defer k.l.Unlock()

	kl, ok := k.m[key]
	if !ok {
		kl = k.p.Get().(*nRWMutex)
		k.m[key] = kl
	}

	ok = kl.lock.TryRLock()
	if ok {
		kl.n++
	}
	return
}

func (k *Krwmutex) Unlock(key any) {
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

type rlocker struct {
	key any
	*Krwmutex
}

func (r *rlocker) Lock()   { (*Krwmutex)(r.Krwmutex).RLock(r.key) }
func (r *rlocker) Unlock() { (*Krwmutex)(r.Krwmutex).RUnlock(r.key) }

func (k *Krwmutex) RLocker(key any) sync.Locker {
	return &rlocker{
		key:      key,
		Krwmutex: k,
	}
}
