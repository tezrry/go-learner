package metric

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"sync/atomic"
)

type QPS struct {
	idx       atomic.Int64
	num       [2]atomic.Int64
	startTime int64
	gauge     prometheus.Gauge
}

func NewQPS(name string) *QPS {
	return &QPS{
		gauge: promauto.NewGauge(prometheus.GaugeOpts{Name: name}),
	}
}

func (q *QPS) Incr(v int64) {
	q.num[q.idx.Load()].Add(v)
}

func (q *QPS) Collect(currTime int64) float64 {
	if q.startTime == 0 {
		//q.num[q.idx.Load()].Store(0)
		v := q.num[q.idx.Load()].Swap(0)
		q.startTime = currTime
		return float64(v)
	}

	span := currTime - q.startTime
	if span <= 0 {
		return 0
	}

	idx := q.idx.Load()
	q.idx.Store((idx + 1) & 1)

	num := q.num[idx].Swap(0)
	q.startTime = currTime

	v := float64(num) / float64(span)
	q.gauge.Set(v)
	return v
}
