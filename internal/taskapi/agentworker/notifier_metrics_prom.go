package agentworker

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"github.com/AlexsanderHamir/Hamix/pkgs/tasks/calltrace"
	"github.com/prometheus/client_golang/prometheus"
)

var registerNotifierMetrics sync.Once

type prometheusNotifierMetrics struct {
	dropped *prometheus.CounterVec
}

//funclogmeasure:skip category=hot-path reason="Prometheus counter increment; drop events trace at notifier publish boundary."
func (m *prometheusNotifierMetrics) RecordNotifierDropped(kind string) {
	if m == nil || m.dropped == nil {
		return
	}
	m.dropped.WithLabelValues(kind).Inc()
}

// RegisterNotifierMetrics registers hamix_agent_notifier_dropped_total once and
// returns a NotifierMetrics for SSE adapters. Safe to call when the worker is
// disabled — observations are no-ops if registration failed.
//
//funclogmeasure:skip category=hot-path reason="One-time Prometheus registration; inner Do callback emits trace lines."
func RegisterNotifierMetrics() NotifierMetrics {
	var out NotifierMetrics
	registerNotifierMetrics.Do(func() {
		counter := prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "hamix",
			Name:      "agent_notifier_dropped_total",
			Help:      "Harness notifier events dropped due to back-pressure (queue full or slow publish).",
		}, []string{"kind"})
		if err := prometheus.Register(counter); err != nil {
			var dup prometheus.AlreadyRegisteredError
			if errors.As(err, &dup) {
				if existing, ok := dup.ExistingCollector.(*prometheus.CounterVec); ok {
					out = &prometheusNotifierMetrics{dropped: existing}
					return
				}
			}
			slog.Warn("prometheus agent notifier metrics register failed",
				"cmd", calltrace.LogCmd, "operation", "taskapi.RegisterNotifierMetrics", "err", err)
			return
		}
		out = &prometheusNotifierMetrics{dropped: counter}
		slog.Info("prometheus agent notifier metrics registered",
			"cmd", calltrace.LogCmd, "operation", "taskapi.RegisterNotifierMetrics")
	})
	return out
}

// RegisterNotifierMetricsOn registers notifier metrics on reg for tests.
//
//funclogmeasure:skip category=tool-required-noop reason="Test-only Prometheus registration helper."
func RegisterNotifierMetricsOn(reg prometheus.Registerer) (NotifierMetrics, error) {
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "hamix",
		Name:      "agent_notifier_dropped_total",
		Help:      "Harness notifier events dropped due to back-pressure (queue full or slow publish).",
	}, []string{"kind"})
	if err := reg.Register(counter); err != nil {
		return nil, fmt.Errorf("register hamix_agent_notifier_dropped_total: %w", err)
	}
	return &prometheusNotifierMetrics{dropped: counter}, nil
}
