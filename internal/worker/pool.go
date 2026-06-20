package worker

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Job describes a document to be processed by the worker pool.
type Job struct {
	DocID       string
	Filename    string
	FileSize    int64
	StoragePath string
}

// ErrPoolStopped is returned when attempting to enqueue after the pool has stopped.
var ErrPoolStopped = fmt.Errorf("worker pool: stopped")

// Pool is a bounded, concurrent worker pool for processing documents asynchronously.
type Pool struct {
	jobs    chan Job
	workers int
	process func(context.Context, Job) error
	ctx     context.Context
	cancel  context.CancelFunc
	started bool
	mu      sync.Mutex
	wg      sync.WaitGroup
}

// New creates a worker pool with n workers and a job buffer of buf.
func New(n int, buf int, process func(context.Context, Job) error) *Pool {
	return &Pool{
		jobs:    make(chan Job, buf),
		workers: n,
		process: process,
	}
}

// Start launches the worker goroutines. Must be called once before Enqueue.
func (p *Pool) Start() {
	p.mu.Lock()
	if p.started {
		p.mu.Unlock()
		panic("worker pool: already started")
	}
	p.started = true
	p.ctx, p.cancel = context.WithCancel(context.Background())
	p.mu.Unlock()

	for range p.workers {
		p.wg.Add(1)
		go p.worker()
	}
}

func (p *Pool) worker() {
	defer p.wg.Done()
	for {
		select {
		case job, ok := <-p.jobs:
			if !ok {
				return
			}
			if err := p.process(p.ctx, job); err != nil {
				slog.Error("worker: job failed", "error", err)
			}
		case <-p.ctx.Done():
			return
		}
	}
}

// Enqueue adds a job to the work queue. Returns ErrPoolStopped if the pool is shut down.
func (p *Pool) Enqueue(job Job) error {
	p.mu.Lock()
	ctx := p.ctx
	if !p.started {
		p.mu.Unlock()
		if ctx != nil && ctx.Err() != nil {
			return ErrPoolStopped
		}
		return fmt.Errorf("worker pool: not started")
	}
	p.mu.Unlock()

	select {
	case p.jobs <- job:
		return nil
	case <-ctx.Done():
		return ErrPoolStopped
	}
}

// Stop signals all workers to shut down and waits for them to finish.
func (p *Pool) Stop() {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		return
	}
	p.started = false
	p.cancel()
	p.mu.Unlock()

	p.wg.Wait()
}
