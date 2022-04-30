package tracing

import (
	"sync"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
)

type Duration struct {
	At       time.Time `json:"at,omitempty"`
	Till     time.Time `json:"till,omitempty"`
	Duration string    `json:"duration,omitempty"`
}
type Function struct {
	Entrypoint       string `json:"entrypoint,omitempty"`
	FilePath         string `json:"filePath,omitempty"`
	Dir              string `json:"dir,omitempty"`
	ProxyServiceAddr string `json:"proxyServiceAddr,omitempty"`
	Url              string `json:"url,omitempty"`
	Status           string `json:"status,omitempty"`
}
type Metric struct {
	Function    Function   `json:"function"`
	Invocations int        `json:"invocations"`
	Success     int        `json:"success"`
	Error       int        `json:"error"`
	Duration    []Duration `json:"duration"`
}

type UselessMetrics struct {
	sync.RWMutex
	fns map[string]*Metric
}

func (mm *UselessMetrics) GetAll() []Metric {
	var m = make([]Metric, len(mm.fns))
	mm.RLock()
	defer mm.RUnlock()
	for _, metric := range mm.fns {
		m = append(m, *metric)
	}
	return m
}

func Manager() UselessMetrics {
	mm := UselessMetrics{}
	mm.RWMutex = sync.RWMutex{}
	mm.fns = make(map[string]*Metric)
	return mm
}

func newMetric(function *deploy.Function) *Metric {
	var f Function
	if function != nil {
		f = copyFunction(function)
	}
	m := Metric{
		Function:    f,
		Invocations: 0,
		Success:     0,
		Error:       0,
		Duration:    []Duration{},
	}
	return &m
}

func copyFunction(function *deploy.Function) Function {
	f := new(Function)
	f.Entrypoint = function.Entrypoint
	f.FilePath = function.FilePath
	f.Dir = function.Dir
	f.ProxyServiceAddr = function.ProxyServiceAddr
	f.Url = function.Url
	f.Status = function.Status
	return *f
}
func (mm *UselessMetrics) Register(fn *deploy.Function) {
	mm.Lock()
	defer mm.Unlock()
	mm.fns[fn.Entrypoint] = newMetric(fn)
}
func (mm *UselessMetrics) Invoked(name string) chan<- bool {
	mm.RLock()
	if mm.fns[name] == nil {
		mm.RUnlock()
		return nil
	}
	mt := mm.fns[name]
	mm.RUnlock()

	d := Duration{
		At:   time.Now(),
		Till: time.Time{},
	}
	mt.Invocations++
	result := make(chan bool)
	go func() {
		if success, _ := <-result; success == true {
			mt.Success++
		} else {
			mt.Error++
		}
		d.Till = time.Now()
		d.Duration = time.Since(d.Till).String()
		mt.Duration = append(mt.Duration, d)

		mm.Lock()
		mm.fns[name] = mt
		mm.Unlock()
	}()
	return result
}
