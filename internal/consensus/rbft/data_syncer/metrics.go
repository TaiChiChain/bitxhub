package data_syncer

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	processEventDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "data_syncer",
			Name:      "process_event_duration_seconds",
			Help:      "the duration of process event",
			Buckets:   prometheus.ExponentialBuckets(0.0001, 2, 10),
		},
		[]string{"event"},
	)
)

func init() {
	prometheus.MustRegister(processEventDuration)
}

func traceProcessEvent(event string, duration time.Duration) {
	processEventDuration.With(prometheus.Labels{"event": event}).Observe(duration.Seconds())
	processEventDuration.With(prometheus.Labels{"event": "all"}).Observe(duration.Seconds())
}
