package shell

import (
	"sync"
	"time"

	"github.com/patrickmn/go-cache"
)

const (
	ScaleUp         = 5
	ScaleDown       = 1
	DefaultDuration = 30 * time.Second
)

var cc = cache.New(DefaultDuration, 1*time.Minute)

type processCache struct {
	scaleInterval time.Duration
	processes     map[string]*processMeta
	Sync          sync.RWMutex
}
type processMeta struct {
	name      string
	instances int
	opts      []Option
	refreshAt time.Time
	shells    map[int]Shell
}

func NewProcessCache(scaleInterval time.Duration) processCache {
	return processCache{
		scaleInterval: scaleInterval,
		processes:     make(map[string]*processMeta),
		Sync:          sync.RWMutex{},
	}
}

func (c *processCache) scale(name string, to int) {
	c.Sync.Lock()
	defer c.Sync.Unlock()
	if c.processes[name].instances < to {
		for i := 1; i <= to-c.processes[name].instances; i++ {
			sh := New(c.processes[name].opts...)
			c.processes[name].shells[i] = sh
			//sh.Run()
			//cc.Increment(name)
			//cc.

			go func(sh Shell) {

			}(sh)
		}
	} else {
		for i := c.processes[name].instances; i >= to; i-- {
			//c.processes[name].shells[i].Kill()
			delete(c.processes[name].shells, i)
		}
	}
}
func (c *processCache) Set(name string) {
	c.Sync.Lock()
	defer c.Sync.Unlock()

	if c.processes[name] == nil {
		c.processes[name] = new(processMeta)
	}
	c.processes[name].refreshAt = time.Now()
	cc.Set(name, time.Now(), DefaultDuration)
}
func (c *processCache) Get(name string) {
	c.Sync.Lock()
	defer c.Sync.Unlock()

	c.processes[name].refreshAt = time.Now()

	if _, hit := cc.Get(name); !hit {
		cc.Set(name, time.Now(), DefaultDuration)
		return
	}

	cc.Replace(name, time.Now(), DefaultDuration)
}
