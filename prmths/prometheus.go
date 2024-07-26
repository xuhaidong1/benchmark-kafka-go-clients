package prmths

import (
	"github.com/prometheus/client_golang/prometheus"
)

var ProduceMessageLatency = prometheus.NewSummaryVec(
	prometheus.SummaryOpts{
		Name: "produce_message_latency",
		Help: "produce_message_latency",
	},
	[]string{"library", "message_size"},
)

func init() {
	prometheus.MustRegister(ProduceMessageLatency)
}
