package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	opsProcessed *prometheus.CounterVec
}

func InitPrometheus() *PrometheusMetrics {
	opsProcessed := promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "function_invocations",
		Help: "The total number of invocation of a deployed function",
	}, []string{"entrypoint", "agent", "statusCode"})
	return &PrometheusMetrics{
		opsProcessed: opsProcessed,
	}
}
func (lm *PrometheusMetrics) Update(name string, agent string, success int) {
	lm.opsProcessed.With(prometheus.Labels{"entrypoint": name, "agent": agent, "statusCode": strconv.Itoa(success)}).Inc()
}

// rate(function_invocations{}[$__rate_interval])
