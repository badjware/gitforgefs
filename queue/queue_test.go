package queue_test

import (
	"log/slog"
	"sync/atomic"
	"testing"
	"time"

	"github.com/badjware/gitforgefs/queue"
)

func waitForInt32(ptr *int32, want int32, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(ptr) == want {
			return true
		}
		time.Sleep(5 * time.Millisecond)
	}
	return atomic.LoadInt32(ptr) == want
}

func TestNewMemoryQueueRunsTasks(t *testing.T) {
	q := queue.NewMemoryQueue(slog.Default(), "test-run", 2, 100)
	defer q.Shutdown()

	var counter int32
	const n = 50

	for i := 0; i < n; i++ {
		err := q.AddTask(func() {
			atomic.AddInt32(&counter, 1)
		})
		if err != nil {
			t.Fatalf("AddTask returned error: %v", err)
		}
	}

	if !waitForInt32(&counter, n, 2*time.Second) {
		t.Fatalf("expected %d tasks to run, got %d", n, atomic.LoadInt32(&counter))
	}
}

func TestNewMemoryQueueBufferFull(t *testing.T) {
	q := queue.NewMemoryQueue(slog.Default(), "test-buffer", 1, 1)
	defer q.Shutdown()

	blockCh := make(chan struct{})
	var ran int32

	// First task: blocks the single worker
	if err := q.AddTask(func() {
		atomic.AddInt32(&ran, 1)
		<-blockCh
	}); err != nil {
		t.Fatalf("first AddTask failed: %v", err)
	}

	// Give worker a moment to pick up the first task
	time.Sleep(50 * time.Millisecond)

	// Second task: should be accepted into the buffer
	if err := q.AddTask(func() {
		atomic.AddInt32(&ran, 1)
	}); err != nil {
		t.Fatalf("second AddTask failed (should fit in buffer): %v", err)
	}

	// Third task: buffer is full and worker is busy, should return error
	if err := q.AddTask(func() {}); err == nil {
		t.Fatalf("third AddTask succeeded but expected error because queue should be full")
	}

	// Unblock first task so worker can finish and run buffered task
	close(blockCh)

	if !waitForInt32(&ran, 2, 2*time.Second) {
		t.Fatalf("expected 2 tasks to run after unblocking, got %d", atomic.LoadInt32(&ran))
	}
}

func TestNewMemoryQueueShutdown(t *testing.T) {
	q := queue.NewMemoryQueue(slog.Default(), "test-shutdown", 1, 1)
	// Shutdown immediately
	q.Shutdown()

	// After shutdown, AddTask should return an error
	if err := q.AddTask(func() {}); err == nil {
		t.Fatalf("AddTask succeeded after Shutdown; expected error")
	}
}
