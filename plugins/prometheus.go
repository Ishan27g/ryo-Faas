package plugins

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Prometheus *metrics

type metrics struct {
	opsProcessed *prometheus.CounterVec
}

func InitPrometheus() {
	opsProcessed := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "function_invocations",
		Help: "The total number of invocation of a deployed function",
	}, []string{"entrypoint", "agent", "statusCode"})
	Prometheus = &metrics{
		opsProcessed: opsProcessed,
	}
}
func (lm *metrics) Update(name string, agent string, success int) {
	// rate(function_invocations{}[$__rate_interval])
	lm.opsProcessed.With(prometheus.Labels{"entrypoint": name, "agent": agent, "statusCode": strconv.Itoa(success)}).Inc()
}
