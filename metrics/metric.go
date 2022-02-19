package metrics

import (
	"fmt"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/proto"
)

type Duration struct {
	At       time.Time `json:"at,omitempty"`
	Till     time.Time `json:"till,omitempty"`
	Duration string    `json:"duration,omitempty"`
}
type Function struct {
	Entrypoint string `json:"entrypoint,omitempty"`
	// file name
	FilePath string `json:"filePath,omitempty"`
	// path to package-dir
	Dir string `json:"dir,omitempty"`
	Zip string `json:"zip,omitempty"`
	// address of agent that manages function
	AtAgent string `json:"atAgent,omitempty"`
	// address of service running on agent
	ProxyServiceAddr string `json:"proxyServiceAddr,omitempty"`
	// function endpoint
	Url    string `json:"url,omitempty"`
	Status string `json:"status,omitempty"`
}
type Metric struct {
	Function    Function   `json:"function"`
	Invocations int        `json:"invocations"`
	Success     int        `json:"success"`
	Error       int        `json:"error"`
	Duration    []Duration `json:"duration"`
}
type Functions struct {
	Fns map[string]*Metric
}

var methods map[string]*Functions

type MetricManager struct{}

const DEPLOY = "Deploy"
const LIST = "List"
const LOGS = "Logs"
const STOP = "Stop"
const DETAILS = "Details"
const UPLOAD = "Upload"

func (mm *MetricManager) GetAll() map[string]*Functions {
	rsp := make(map[string]*Functions)
	for s, functions := range methods {
		for _, m := range functions.Fns {
			if m.Invocations != 0 {
				rsp[s] = functions
			}
		}
	}
	return rsp
}
func (mm *MetricManager) Deployed(fn *deploy.Function) chan<- bool {
	mm.check(fn)
	return methods[DEPLOY].Invoked(fn.Entrypoint)
}

func (mm *MetricManager) check(fn *deploy.Function) {
	if methods[DEPLOY].Fns[fn.Entrypoint] == nil {
		methods[DEPLOY].Fns[fn.Entrypoint] = newMetric(fn)
		methods[LIST].Fns[fn.Entrypoint] = newMetric(fn)
		methods[LOGS].Fns[fn.Entrypoint] = newMetric(fn)
		methods[STOP].Fns[fn.Entrypoint] = newMetric(fn)
		methods[DETAILS].Fns[fn.Entrypoint] = newMetric(fn)
		methods[UPLOAD].Fns[fn.Entrypoint] = newMetric(fn)
	}
}
func (mm *MetricManager) List(entrypoint string) chan<- bool {
	return methods[LIST].Invoked(entrypoint)
}
func (mm *MetricManager) Logs(entrypoint string) chan<- bool {
	return methods[LOGS].Invoked(entrypoint)
}
func (mm *MetricManager) Stop(entrypoint string) chan<- bool {
	return methods[STOP].Invoked(entrypoint)
}
func (mm *MetricManager) Details(entrypoint string) chan<- bool {
	return methods[DETAILS].Invoked(entrypoint)
}
func (mm *MetricManager) Upload(fn *deploy.Function) chan<- bool {
	mm.check(fn)
	return methods[UPLOAD].Invoked(fn.Entrypoint)
}
func Manager() MetricManager {
	m := MetricManager{}
	methods = make(map[string]*Functions)
	methods[DEPLOY] = NewMetricMap()
	methods[LIST] = NewMetricMap()
	methods[LOGS] = NewMetricMap()
	methods[STOP] = NewMetricMap()
	methods[DETAILS] = NewMetricMap()
	methods[UPLOAD] = NewMetricMap()
	return m
}

func NewMetricMap() *Functions {
	return &Functions{Fns: make(map[string]*Metric)}
}
func newMetric(function *deploy.Function) *Metric {
	var f Function
	if function != nil {
		f = copy(function)
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

func copy(function *deploy.Function) Function {
	f := new(Function)
	f.Entrypoint = function.Entrypoint
	f.FilePath = function.FilePath
	f.Dir = function.Dir
	f.Zip = function.Zip
	f.AtAgent = function.AtAgent
	f.ProxyServiceAddr = function.ProxyServiceAddr
	f.Url = function.Url
	f.Status = function.Status
	return *f
}
func (m *Functions) Invoked(name string) chan<- bool {
	if m.Fns[name] == nil {
		m.Fns[name] = newMetric(nil)
	}
	mt := m.Fns[name]
	// d := mt.Duration[len(mt.Duration)-1]
	d := Duration{
		At:   time.Now(),
		Till: time.Time{},
	}
	fmt.Println("entrypoint - ", name)
	mt.Invocations++
	result := make(chan bool)
	go func(m *Functions) {
		if success, _ := <-result; success == true {
			mt.Success++
		} else {
			mt.Error++
		}
		d.Till = time.Now()
		d.Duration = time.Since(d.Till).String()
		mt.Duration = append(mt.Duration, d)
		m.Fns[name] = mt
	}(m)
	return result
}
