package worker

import (
	"context"
	"fmt"
	"sync"
)

type Job struct {
	DocID       string
	Filename    string
	FileSize    int64
	StoragePath string
}

var ErrPoolStopped = fmt.Errorf("worker pool: stopped")

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

func New(n int, buf int, process func(context.Context, Job) error) *Pool {
	return &Pool{
		jobs:    make(chan Job, buf),
		workers: n,
		process: process,
	}
}

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
			p.process(p.ctx, job)
		case <-p.ctx.Done():
			return
		}
	}
}

func (p *Pool) Enqueue(job Job) error {
	p.mu.Lock()
	if !p.started {
		p.mu.Unlock()
		if p.ctx != nil && p.ctx.Err() != nil {
			return ErrPoolStopped
		}
		return fmt.Errorf("worker pool: not started")
	}
	p.mu.Unlock()

	select {
	case p.jobs <- job:
		return nil
	case <-p.ctx.Done():
		return ErrPoolStopped
	}
}

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
