package ksync

import (
	"sync"
)

type Krwmutex struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*krwlock
}

type krwlock struct {
	lock *sync.RWMutex
	n    uint64
}

func DefaultKrwmutex() *Krwmutex {
	return &Krwmutex{
		l: &sync.Mutex{},
		p: &sync.Pool{
			New: func() any {
				return &krwlock{
					lock: &sync.RWMutex{},
				}
			},
		},
		m: make(map[any]*krwlock),
	}
}

func NewKrwmutex(locker ...sync.Locker) *Kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
	}
	return km
}

func (krw *Krwmutex) Lock(key any) {
	krw.l.Lock()

	kl, ok := krw.m[key]
	if !ok {
		kl = krw.p.Get().(*krwlock)
		krw.m[key] = kl
	}
	kl.n++
	krw.l.Unlock()

	kl.lock.Lock()
}

func (krw *Krwmutex) RLock(key any) {
	krw.l.Lock()

	kl, ok := krw.m[key]
	if !ok {
		kl = krw.p.Get().(*krwlock)
		krw.m[key] = kl
	}
	kl.n++
	krw.l.Unlock()

	kl.lock.RLock()
}

func (krw *Krwmutex) RUnlock(key any) {
	krw.l.Lock()
	defer krw.l.Unlock()

	kl, ok := krw.m[key]
	if !ok {
		return
	}

	kl.n--
	if kl.n == 0 {
		krw.p.Put(kl)
		delete(krw.m, key)
	}

	kl.lock.RUnlock()
}

func (krw *Krwmutex) TryLock(key any) (ok bool) {
	krw.l.Lock()
	defer krw.l.Unlock()

	kl, ok := krw.m[key]
	if !ok {
		kl = krw.p.Get().(*krwlock)
		krw.m[key] = kl
	}

	ok = kl.lock.TryLock()
	if ok {
		kl.n++
	}
	return
}

func (krw *Krwmutex) TryRLock(key any) (ok bool) {
	krw.l.Lock()
	defer krw.l.Unlock()

	kl, ok := krw.m[key]
	if !ok {
		kl = krw.p.Get().(*krwlock)
		krw.m[key] = kl
	}

	ok = kl.lock.TryRLock()
	if ok {
		kl.n++
	}
	return
}

func (krw *Krwmutex) Unlock(key any) {
	krw.l.Lock()
	defer krw.l.Unlock()

	kl, ok := krw.m[key]
	if !ok {
		return
	}

	kl.n--
	if kl.n == 0 {
		krw.p.Put(kl)
		delete(krw.m, key)
	}

	kl.lock.Unlock()
}

type rlocker struct {
	key any
	*Krwmutex
}

func (r *rlocker) Lock()   { (*Krwmutex)(r.Krwmutex).RLock(r.key) }
func (r *rlocker) Unlock() { (*Krwmutex)(r.Krwmutex).RUnlock(r.key) }

func (krw *Krwmutex) RLocker(key any) sync.Locker {
	return &rlocker{
		key:      key,
		Krwmutex: krw,
	}
}
