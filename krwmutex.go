package kmutex

import "sync"

type krwmutex struct {
	l sync.Locker
	m map[any]*krwlock
}

type krwlock struct {
	lock *sync.RWMutex
	ln   uint64
	rwln uint64
}

func DefaultKrwmutex() *krwmutex {
	return &krwmutex{
		l: &sync.Mutex{},
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
		kl = &krwlock{
			lock: &sync.RWMutex{},
		}
		krw.m[key] = kl
	}
	kl.ln++
	krw.l.Unlock()

	kl.lock.Lock()
}

func (krw *krwmutex) RLock(key any) {
	krw.l.Lock()
	kl, ok := krw.m[key]
	if !ok {
		kl = &krwlock{
			lock: &sync.RWMutex{},
		}
		krw.m[key] = kl
	}
	kl.rwln++
	krw.l.Unlock()

	kl.lock.RLock()
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

func (krw *krwmutex) RUnlock(key any) {
	krw.l.Lock()
	defer krw.l.Unlock()
	kl, ok := krw.m[key]
	if ok {
		kl.rwln--
		if kl.ln == 0 && kl.rwln == 0 {
			delete(krw.m, key)
		}
		kl.lock.RUnlock()
	}
}

func (krw *krwmutex) TryLock(key any) (ok bool) {
	krw.l.Lock()
	defer krw.l.Unlock()
	kl, ok := krw.m[key]
	if !ok {
		kl = &krwlock{
			lock: &sync.RWMutex{},
		}
		krw.m[key] = kl
	}
	ok = kl.lock.TryLock()
	if ok {
		kl.ln++
	}
	return
}

func (krw *krwmutex) TryRLock(key any) (ok bool) {
	krw.l.Lock()
	defer krw.l.Unlock()
	kl, ok := krw.m[key]
	if !ok {
		kl = &krwlock{
			lock: &sync.RWMutex{},
		}
		krw.m[key] = kl
	}
	ok = kl.lock.TryRLock()
	if ok {
		kl.rwln++
	}
	return
}

func (krw *krwmutex) Unlock(key any) {
	krw.l.Lock()
	defer krw.l.Unlock()
	kl, ok := krw.m[key]
	if ok {
		kl.ln--
		if kl.ln == 0 && kl.rwln == 0 {
			delete(krw.m, key)
		}
		kl.lock.Unlock()
	}
}
