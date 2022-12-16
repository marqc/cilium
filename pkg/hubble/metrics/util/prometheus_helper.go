package util

import (
	"time"

	"github.com/cilium/cilium/pkg/lock"
	"github.com/prometheus/client_golang/prometheus"
)

type cacheRecord struct {
	timestamp time.Time
	labels    []string
}

// CounterVec is a wrapper for prometheus.CounterVec that keeps track of last
// entry access and removes data series that haven't been used within a specified TTL.
type CounterVec struct {
	*prometheus.CounterVec
	ttl      time.Duration
	cache    map[string]cacheRecord
	gcTicker *time.Ticker
	mutex    *lock.Mutex
}

// NewCounterVec creates a CounterVec wrapper instance with default gc
// reconciliation interval of 1 minute.
func NewCounterVec(opts prometheus.CounterOpts, labels []string, ttl time.Duration) *CounterVec {
	return NewTTLCounterVecWithReconciliation(opts, labels, ttl, time.Minute)
}

// NewTTLCounterVecWithReconciliation creates a CounterVec wrapper instance with
// specified ttl and gc reconciliation interval.
func NewTTLCounterVecWithReconciliation(opts prometheus.CounterOpts, labels []string, ttl time.Duration, reconciliation time.Duration) *CounterVec {
	var ticker *time.Ticker
	if ttl > 0 {
		ticker = time.NewTicker(reconciliation)
	}
	counter := &CounterVec{
		prometheus.NewCounterVec(opts, labels),
		ttl,
		make(map[string]cacheRecord),
		ticker,
		&lock.Mutex{},
	}

	if nil != ticker {
		go counter.gc()
	}
	return counter
}

// WithLabelValues updates given labels set access time and returns
// prometheus.Counter.
func (v *CounterVec) WithLabelValues(lvs ...string) prometheus.Counter {
	if v.ttl > 0 {
		v.mutex.Lock()
		defer v.mutex.Unlock()
		cacheKey := ""
		for _, l := range lvs {
			cacheKey += l + "|"
		}
		v.cache[cacheKey] = cacheRecord{time.Now(), lvs}
	}
	return v.CounterVec.WithLabelValues(lvs...)
}

// gc removes data series that exceed ttl.
func (v *CounterVec) gc() {
	for _ = range v.gcTicker.C {
		v.mutex.Lock()
		for key, cacheRecord := range v.cache {
			if cacheRecord.timestamp.Add(v.ttl).Before(time.Now()) {
				v.DeleteLabelValues(cacheRecord.labels...)
				delete(v.cache, key)
			}
		}

		v.mutex.Unlock()
	}
}

// HistogramVec is a wrapper for prometheus.HistogramVec that keeps track of last
// entry access and removes data series that haven't been used within a specified TTL.
type HistogramVec struct {
	*prometheus.HistogramVec
	ttl      time.Duration
	cache    map[string]cacheRecord
	gcTicker *time.Ticker
	mutex    lock.Mutex
}

// NewHistogramVec creates a HistogramVec wrapper instance with default gc
// reconciliation interval of 1 minute.
func NewHistogramVec(opts prometheus.HistogramOpts, labels []string, ttl time.Duration) *HistogramVec {
	return NewTTLHistogramVecWithReconciliation(opts, labels, ttl, time.Minute)
}

// NewTTLHistogramVecWithReconciliation creates a HistogramVec wrapper instance with
// specified ttl and gc reconciliation interval.
func NewTTLHistogramVecWithReconciliation(opts prometheus.HistogramOpts, labels []string, ttl time.Duration, reconciliation time.Duration) *HistogramVec {
	var ticker *time.Ticker
	if ttl > 0 {
		ticker = time.NewTicker(reconciliation)
	}
	counter := &HistogramVec{
		prometheus.NewHistogramVec(opts, labels),
		ttl,
		make(map[string]cacheRecord),
		ticker,
		lock.Mutex{},
	}

	if nil != ticker {
		go counter.gc()
	}
	return counter
}

// WithLabelValues updates given labels set access time and returns
// prometheus.Observer.
func (v *HistogramVec) WithLabelValues(lvs ...string) prometheus.Observer {
	if v.ttl > 0 {
		v.mutex.Lock()
		defer v.mutex.Unlock()
		cacheKey := ""
		for _, l := range lvs {
			cacheKey += l + "|"
		}
		v.cache[cacheKey] = cacheRecord{time.Now(), lvs}
	}
	return v.HistogramVec.WithLabelValues(lvs...)
}

// gc removes data series that exceed ttl.
func (v *HistogramVec) gc() {
	for _ = range v.gcTicker.C {
		v.mutex.Lock()

		for key, cacheRecord := range v.cache {
			if cacheRecord.timestamp.Add(v.ttl).Before(time.Now()) {
				v.DeleteLabelValues(cacheRecord.labels...)
				delete(v.cache, key)
			}
		}

		v.mutex.Unlock()
	}
}
