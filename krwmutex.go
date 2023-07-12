package ksync

import (
	"sync"
)

type krwmutex struct {
	l sync.Locker
	p *sync.Pool
	m map[any]*krwlock
}

type krwlock struct {
	lock *sync.RWMutex
	n    uint64
}

func DefaultKrwmutex() *krwmutex {
	return &krwmutex{
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

func NewKrwmutex(locker ...sync.Locker) *kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
	}
	return km
}

func (krw *krwmutex) Lock(key any) {
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

func (krw *krwmutex) RLock(key any) {
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

func (krw *krwmutex) RUnlock(key any) {
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

func (krw *krwmutex) TryLock(key any) (ok bool) {
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

func (krw *krwmutex) TryRLock(key any) (ok bool) {
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

func (krw *krwmutex) Unlock(key any) {
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
	*krwmutex
}

func (r *rlocker) Lock()   { (*krwmutex)(r.krwmutex).RLock(r.key) }
func (r *rlocker) Unlock() { (*krwmutex)(r.krwmutex).RUnlock(r.key) }

func (krw *krwmutex) RLocker(key any) sync.Locker {
	return &rlocker{
		key:      key,
		krwmutex: krw,
	}
}
