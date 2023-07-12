package ksync

import (
	"fmt"
	"testing"
)

func TestKMutex(t *testing.T) {
	km := DefaultKmutex()
	km.Lock("test")
	km.Unlock("test")

	km.Lock("test")
	fmt.Printf("try lock: %v\n", km.TryLock("test"))
	km.Unlock("test")
	fmt.Printf("try lock: %v\n", km.TryLock("test"))
}
