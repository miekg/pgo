package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CmdCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pgo",
		Subsystem: "command",
		Name:      "count",
		Help:      "Counter for the number of commands executed.",
	}, []string{"service", "cmd", "subcmd"})

	CmdErrorCount = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "pgo",
		Subsystem: "command",
		Name:      "error_count",
		Help:      "Counter for the number of commands executed that resulted in an error",
	}, []string{"service", "cmd", "subcmd"})
)
