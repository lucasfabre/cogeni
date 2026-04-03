package processor

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/lucasfabre/cogeni/src/config"
)

func TestRuntimePool_EnforcesLimit(t *testing.T) {
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	pool := NewRuntimePool(cfg, 2) // Limit of 2

	var acquired int32
	var finished int32

	// Acquire two tokens immediately
	rt1, err := pool.Acquire()
	if err != nil {
		t.Fatalf("Failed to acquire rt1: %v", err)
	}
	atomic.AddInt32(&acquired, 1)

	rt2, err := pool.Acquire()
	if err != nil {
		t.Fatalf("Failed to acquire rt2: %v", err)
	}
	atomic.AddInt32(&acquired, 1)

	// Attempt a third acquisition in a goroutine
	done := make(chan struct{})
	started := make(chan struct{})
	go func() {
		close(started)
		rt3, _ := pool.Acquire()
		atomic.AddInt32(&acquired, 1)
		pool.Release(rt3)
		atomic.AddInt32(&finished, 1)
		close(done)
	}()

	// Wait for goroutine to start
	<-started

	// Non-blocking check to ensure the channel hasn't received yet
	// because the semaphore should be full.
	select {
	case <-done:
		t.Fatal("Third runtime was acquired without blocking, limit enforcement failed")
	case <-time.After(50 * time.Millisecond): // Deterministic enough for Go runtime scheduling check
		if atomic.LoadInt32(&acquired) != 2 {
			t.Errorf("Expected exactly 2 acquisitions so far, got %d", atomic.LoadInt32(&acquired))
		}
	}

	// Release one runtime to unblock the goroutine
	pool.Release(rt1)

	// Wait for the goroutine to finish
	<-done

	if atomic.LoadInt32(&acquired) != 3 {
		t.Errorf("Expected 3 total acquisitions after unblocking, got %d", atomic.LoadInt32(&acquired))
	}
	if atomic.LoadInt32(&finished) != 1 {
		t.Errorf("Expected 1 finished runtime, got %d", atomic.LoadInt32(&finished))
	}

	pool.Release(rt2)
}
