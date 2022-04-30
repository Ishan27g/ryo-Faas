package metric

import (
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
)

var m metricMmmm

type Monitor interface {
	Invoked(m tracing.Metric)
}
type invocation struct {
	name  string
	count int
}

func (i *invocation) invoked() {
	i.count++
}

type metricMmmm struct {
	l              sync.Mutex
	registered     map[string]string
	currentMetrics map[string]*invocation
}

func (e *metricMmmm) analyse(sc *Scale) {
	e.l.Lock()
	defer e.l.Unlock()
	fmt.Println("scale factors - ")
	var inv []invocation
	for _, invocation := range e.currentMetrics {
		inv = append(inv, *invocation)
		e.currentMetrics[invocation.name].count = 0
	}
	scaled := sc.scale(inv...)
	for name, factor := range scaled {
		fmt.Println(name, factor)
	}
}
func Start() Monitor {
	m = metricMmmm{
		registered:     map[string]string{},
		currentMetrics: make(map[string]*invocation),
		l:              sync.Mutex{},
	}
	sc := Scale{functions: map[string]*int{}}
	go func() {
		for {
			<-time.After(5 * time.Second)
			m.analyse(&sc)
		}
	}()
	return &m
}
func Register(functionName string) {
	m.l.Lock()
	defer m.l.Unlock()
	m.registered[functionName] = functionName
	m.currentMetrics[functionName] = &invocation{
		name:  functionName,
		count: 0,
	}
}

func (e *metricMmmm) Invoked(m tracing.Metric) {
	e.l.Lock()
	defer e.l.Unlock()
	if e.registered[m.Function.Entrypoint] == "" {
		fmt.Println(m.Function.Entrypoint + " not registered")
		return
	}
	(*e.currentMetrics[m.Function.Entrypoint]).invoked()
}

//func Register(functionName string) {
//	m.l.Lock()
//	defer m.l.Unlock()
//	m.registered[functionName] = functionName
//	//defer m.monitor(functionName)
//	// defer i.invoked()
//	*(m.currentMetrics[functionName]) = invocation{
//		name:  functionName,
//		count: 0,
//	}
//	//m.store.On(store.DocumentCREATE, func(document store.Doc) {
//	//	rec, _ := json.Marshal(document.Data.Value)
//	//	var m tracing.Metric
//	//	err := json.Unmarshal(rec, &m)
//	//	if err != nil {
//	//		fmt.Println(err.Error())
//	//		return
//	//	}
//	//	fmt.Println(m)
//	//	i.Metric = m
//	//}, functionName)
//}
//
//func (e *metricMmmm) Invoked(m tracing.Metric) {
//	e.l.Lock()
//	defer e.l.Unlock()
//	if e.registered[m.Function.Entrypoint] == "" {
//		return
//	}
//	defer (*e.currentMetrics[m.Function.Entrypoint]).invoked()
//	//var data map[string]interface{}
//	//rec, _ := json.Marshal(m)
//	//err := json.Unmarshal(rec, &data)
//	//if err != nil {
//	//	fmt.Println(err.Error())
//	//	return
//	//}
//	//e.store.Create(m.Function.Entrypoint, data)
//
//}
////
////func (e *metricMmmm) monitor(functionName string) {
////	e.store.On(store.DocumentUPDATE, func(document store.Doc) {
////
////		// monitor again
////		e.monitor(functionName)
////
////		rec, _ := json.Marshal(document.Data.Value)
////		var m tracing.Metric
////		err := json.Unmarshal(rec, &m)
////		if err != nil {
////			fmt.Println(err.Error())
////			return
////		}
////		fmt.Println(m)
////	}, functionName)
////}
