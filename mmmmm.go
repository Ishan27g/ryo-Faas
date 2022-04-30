package main

import (
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/metric"
	"github.com/Ishan27g/ryo-Faas/pkg/tracing"
)

var mockMetric = func() tracing.Metric {
	return tracing.Metric{
		Function: tracing.Function{
			Entrypoint:       "something",
			FilePath:         "ok",
			Dir:              "okk",
			ProxyServiceAddr: "/123",
			Url:              "/1/2/3",
			Status:           "what",
		},
		Invocations: 2,
		Success:     3,
		Error:       1,
		Duration: []tracing.Duration{
			{
				At:       time.Unix(1, 1),
				Till:     time.Unix(2, 1),
				Duration: "1s",
			},
		},
	}
}

func main() {
	m := metric.Start()
	metric.Register("something")

	mockMetric := mockMetric()

	m.Invoked(mockMetric)
	m.Invoked(mockMetric)
	m.Invoked(mockMetric)
	m.Invoked(mockMetric)
	m.Invoked(mockMetric)
	m.Invoked(mockMetric)

	<-time.After(100 * time.Second)
}
