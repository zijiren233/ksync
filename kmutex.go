package kmutex

import "sync"

type kmutex struct {
	l sync.Locker
	m map[any]*klock
}

type klock struct {
	lock *sync.Mutex
	ln   uint64
}

func DefaultKmutex() *kmutex {
	return &kmutex{
		l: &sync.Mutex{},
		m: make(map[any]*klock),
	}
}

func NewKmutex(locker ...sync.Locker) *kmutex {
	km := DefaultKmutex()
	for _, lock := range locker {
		km.l = lock
	}
	return km
}

func (km *kmutex) Unlock(key any) {
	km.l.Lock()
	defer km.l.Unlock()
	kl, ok := km.m[key]
	if ok {
		kl.ln--
		if kl.ln == 0 {
			delete(km.m, key)
		}
		kl.lock.Unlock()
	}
}

func (km *kmutex) Lock(key any) {
	km.l.Lock()
	kl, ok := km.m[key]
	if !ok {
		kl = &klock{
			ln:   0,
			lock: &sync.Mutex{},
		}
		km.m[key] = kl
	}
	kl.ln++
	km.l.Unlock()

	kl.lock.Lock()
}

func (km *kmutex) TryLock(key any) (ok bool) {
	km.l.Lock()
	defer km.l.Unlock()
	kl, ok := km.m[key]
	if !ok {
		kl = &klock{
			ln:   0,
			lock: &sync.Mutex{},
		}
		km.m[key] = kl
	}
	ok = kl.lock.TryLock()
	if ok {
		kl.ln++
	}
	return
}
