package main

import (
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/metric"
)

func main() {
	m := metric.Start()
	metric.Register("something")

	m.Invoked("something")
	m.Invoked("something")
	m.Invoked("something")
	m.Invoked("something")
	m.Invoked("something")
	m.Invoked("something")

	<-time.After(100 * time.Second)
}
