package draftassistmetrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// Ready reasons for GET /draft-assist/ready when ready=false.
const (
	ReasonNoRunner    = "no_runner"
	ReasonMissingKey  = "missing_key"
	ReasonSidecarDown = "sidecar_down"
)

var (
	registerOnce sync.Once

	runFirstEventMS prometheus.Histogram
	watchdogTotal   prometheus.Counter
)

// RegisterOn registers draft-assist latency and watchdog metrics. Safe to call
// more than once; subsequent calls are no-ops.
//
//funclogmeasure:skip category=hot-path reason="Prometheus registration; no request-scoped operation to trace."
func RegisterOn(reg prometheus.Registerer) {
	registerOnce.Do(func() {
		runFirstEventMS = prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "taskapi_draftassist_run_first_event_ms",
			Help:    "Milliseconds from run accept to first SSE status/token/tool/error event.",
			Buckets: []float64{5, 10, 25, 50, 100, 250, 500, 1000, 2500, 5000},
		})
		watchdogTotal = prometheus.NewCounter(prometheus.CounterOpts{
			Name: "taskapi_draftassist_watchdog_total",
			Help: "Times the SPA/client silence watchdog fired (instrumented by callers).",
		})
		_ = reg.Register(runFirstEventMS)
		_ = reg.Register(watchdogTotal)
	})
}

// ObserveFirstEvent records latency from runStart to now in milliseconds.
//
//funclogmeasure:skip category=hot-path reason="Histogram observe on SSE emit path; caller already traces."
func ObserveFirstEvent(runStart time.Time) {
	if runFirstEventMS == nil || runStart.IsZero() {
		return
	}
	runFirstEventMS.Observe(float64(time.Since(runStart).Milliseconds()))
}

// IncWatchdog increments the watchdog counter.
//
//funclogmeasure:skip category=hot-path reason="Counter increment; caller already traces."
func IncWatchdog() {
	if watchdogTotal == nil {
		return
	}
	watchdogTotal.Inc()
}
