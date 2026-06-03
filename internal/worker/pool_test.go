package worker_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/croko/language-app/internal/worker"
)

func TestStartStop(t *testing.T) {
	p := worker.New(1, 1, func(_ context.Context, _ worker.Job) error {
		return nil
	})
	p.Start()
	p.Stop()
}

func TestStartTwicePanics(t *testing.T) {
	p := worker.New(1, 1, func(_ context.Context, _ worker.Job) error {
		return nil
	})
	p.Start()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on second Start")
		}
	}()
	p.Start()
}

func TestEnqueueBeforeStartReturnsError(t *testing.T) {
	p := worker.New(1, 1, func(_ context.Context, _ worker.Job) error {
		return nil
	})
	err := p.Enqueue(worker.Job{DocID: "1"})
	if err == nil {
		t.Fatal("expected error when enqueuing before Start")
	}
}

func TestEnqueueAfterStopReturnsErrPoolStopped(t *testing.T) {
	p := worker.New(1, 1, func(_ context.Context, _ worker.Job) error {
		return nil
	})
	p.Start()
	p.Stop()
	err := p.Enqueue(worker.Job{DocID: "1"})
	if err != worker.ErrPoolStopped {
		t.Fatalf("expected ErrPoolStopped, got %v", err)
	}
}

func TestProcessCalledWithJob(t *testing.T) {
	got := make(chan worker.Job, 1)
	p := worker.New(1, 1, func(_ context.Context, job worker.Job) error {
		got <- job
		return nil
	})
	p.Start()
	defer p.Stop()

	expected := worker.Job{
		DocID:       "doc-1",
		Filename:    "test.epub",
		FileSize:    1024,
		StoragePath: "/uploads/doc-1.epub",
	}
	p.Enqueue(expected)

	select {
	case j := <-got:
		if j.DocID != expected.DocID {
			t.Fatalf("DocID: got %q, want %q", j.DocID, expected.DocID)
		}
		if j.Filename != expected.Filename {
			t.Fatalf("Filename: got %q, want %q", j.Filename, expected.Filename)
		}
		if j.FileSize != expected.FileSize {
			t.Fatalf("FileSize: got %d, want %d", j.FileSize, expected.FileSize)
		}
		if j.StoragePath != expected.StoragePath {
			t.Fatalf("StoragePath: got %q, want %q", j.StoragePath, expected.StoragePath)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for job to be processed")
	}
}

func TestMultipleJobsProcessed(t *testing.T) {
	const count = 5
	var mu sync.Mutex
	processed := make([]string, 0, count)
	done := make(chan struct{}, count)

	p := worker.New(2, count, func(_ context.Context, job worker.Job) error {
		mu.Lock()
		processed = append(processed, job.DocID)
		mu.Unlock()
		done <- struct{}{}
		return nil
	})
	p.Start()
	defer p.Stop()

	for i := range count {
		p.Enqueue(worker.Job{DocID: string(rune('0' + i))})
	}

	for range count {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for all jobs")
		}
	}

	mu.Lock()
	if len(processed) != count {
		t.Fatalf("processed %d jobs, want %d", len(processed), count)
	}
	mu.Unlock()
}

func TestConcurrentEnqueue(t *testing.T) {
	const total = 20
	var processed atomic.Int64
	done := make(chan struct{}, total)

	p := worker.New(4, total, func(_ context.Context, _ worker.Job) error {
		processed.Add(1)
		done <- struct{}{}
		return nil
	})
	p.Start()
	defer p.Stop()

	var wg sync.WaitGroup
	for range total {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p.Enqueue(worker.Job{DocID: "concurrent"})
		}()
	}
	wg.Wait()

	for range total {
		select {
		case <-done:
		case <-time.After(time.Second):
			t.Fatal("timed out waiting for concurrent jobs")
		}
	}

	if n := processed.Load(); n != total {
		t.Fatalf("processed %d jobs, want %d", n, total)
	}
}
