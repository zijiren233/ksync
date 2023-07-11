package kmutex

import (
	"fmt"
	"testing"
)

func TestKRWMutex(t *testing.T) {
	km := DefaultKrwmutex()
	km.Lock("test")
	km.Unlock("test")

	km.Lock("test")
	fmt.Printf("try lock: %v\n", km.TryLock("test"))
	km.Unlock("test")
	fmt.Printf("try lock: %v\n", km.TryLock("test"))
	km.Unlock("test")

	km.RLock("test")
	fmt.Printf("try rlock: %v\n", km.TryRLock("test"))
	km.RUnlock("test")
	km.RUnlock("test")
	fmt.Printf("try rlock: %v\n", km.TryRLock("test"))
	km.RUnlock("test")

	km.Lock("test")
	fmt.Printf("try rlock: %v\n", km.TryRLock("test"))
	km.Unlock("test")
	fmt.Printf("try rlock: %v\n", km.TryRLock("test"))
	km.RUnlock("test")
}
