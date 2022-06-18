package scale

import (
	"sync"
)

const (
	min = 1
	two = 2
	max = 3
)

func NewMetricsMonitor() *Monitor {
	m := Monitor{
		registered:     map[string]string{},
		currentMetrics: make(map[string]*invocation),
		l:              sync.Mutex{},
	}
	return &m
}

func (e *Monitor) Invoked(entrypoint string) {
	e.l.Lock()
	defer e.l.Unlock()
	if e.registered[entrypoint] == "" {
		e.register(entrypoint)
	}
	(*e.currentMetrics[entrypoint]).invoked()
}

type Monitor struct {
	l              sync.Mutex
	registered     map[string]string
	currentMetrics map[string]*invocation
}

type invocation struct {
	name  string
	count int
}

func (i *invocation) invoked() {
	i.count++
}
func (e *Monitor) getInvocations() []invocation {
	e.l.Lock()
	defer e.l.Unlock()
	var inv []invocation
	for _, invocation := range e.currentMetrics {
		inv = append(inv, *invocation)
		e.currentMetrics[invocation.name].count = 0
	}
	return inv
}

func (e *Monitor) register(entrypoint string) {
	e.registered[entrypoint] = entrypoint
	e.currentMetrics[entrypoint] = &invocation{
		name:  entrypoint,
		count: 0,
	}
}
