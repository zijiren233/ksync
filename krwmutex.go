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
	lock *sync.RWMutex
	n    uint64
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

func NewKrwmutex(locker ...sync.Locker) *Kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
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

	kl.n--
	if kl.n == 0 {
		k.p.Put(kl)
		delete(k.m, key)
	}

	kl.lock.RUnlock()
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

	kl.n--
	if kl.n == 0 {
		k.p.Put(kl)
		delete(k.m, key)
	}

	kl.lock.Unlock()
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
