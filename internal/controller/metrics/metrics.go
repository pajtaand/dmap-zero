package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	// Controller metrics
	RESTHTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint"},
	)

	// Agent metrics
	AgentPresentImagesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_present_images_total",
			Help: "Number of currently present images on agent",
		},
		[]string{"agent"},
	)
	AgentRunningModulesGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "agent_running_modules_total",
			Help: "Number of currently running modules on agent",
		},
		[]string{"agent"},
	)
)

func init() {
	prometheus.MustRegister(RESTHTTPRequestsTotal)
	prometheus.MustRegister(AgentPresentImagesGauge)
	prometheus.MustRegister(AgentRunningModulesGauge)
}
