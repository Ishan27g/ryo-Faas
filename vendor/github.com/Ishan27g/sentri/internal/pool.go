package internal

import (
	"context"

	"github.com/alitto/pond"
)

var p *Pool

const (
	maxWorkers  = 2
	maxCapacity = 5
)

type Pool struct {
	pool *pond.WorkerPool
}

func (p *Pool) Count() uint64 {
	return p.pool.SubmittedTasks()
}
func (p *Pool) Submit(f func()) {
	p.pool.Submit(f)
}
func GetPool() *Pool {
	if p == nil {
		InitPool(context.Background())
	}
	return p
}
func InitPool(ctx context.Context) {
	p = &Pool{}
	// unbuffered (blocking) pool
	p.pool = pond.New(maxWorkers, maxCapacity, pond.MinWorkers(maxWorkers), pond.Context(ctx))

	go func() {
		<-ctx.Done()
		p.pool.StopAndWait()
	}()
}
