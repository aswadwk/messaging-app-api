package services

import (
	"sync"
	"sync/atomic"
)

type WorkerPool struct {
	tasks       chan func()
	wg          sync.WaitGroup
	activeCount int64 // Atomic counter for active workers
	totalCount  int64 // Atomic counter for total workers
}

func NewWorkerPool(workers int) *WorkerPool {
	p := &WorkerPool{
		tasks: make(chan func(), 100),
	}

	atomic.StoreInt64(&p.totalCount, int64(workers))

	for range workers {
		go func() {
			for task := range p.tasks {
				atomic.AddInt64(&p.activeCount, 1)
				task()
				atomic.AddInt64(&p.activeCount, -1)
				p.wg.Done()
			}
		}()
	}

	return p
}

func (p *WorkerPool) Submit(task func()) {
	p.wg.Add(1)
	p.tasks <- task
}

func (p *WorkerPool) Stop() {
	close(p.tasks)
	p.wg.Wait()
}

// GetActiveWorkerCount returns number of currently active workers
func (p *WorkerPool) GetActiveWorkerCount() int64 {
	return atomic.LoadInt64(&p.activeCount)
}

// GetTotalWorkerCount returns total number of workers
func (p *WorkerPool) GetTotalWorkerCount() int64 {
	return atomic.LoadInt64(&p.totalCount)
}
