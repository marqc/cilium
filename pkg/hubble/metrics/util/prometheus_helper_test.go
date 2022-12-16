package util

import (
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrometheusCounterVecWithTtl(t *testing.T) {
	t.Parallel()

	// given
	registry := prometheus.NewRegistry()

	labels := []string{"flag", "family"}

	sut := NewTTLCounterVecWithReconciliation(prometheus.CounterOpts{
		Namespace: "ns",
		Name:      "metric_name",
	}, labels, 10*time.Millisecond, 10*time.Millisecond)
	defer sut.gcTicker.Stop()

	registry.MustRegister(sut)

	// when
	// send metrics each millisecond
	sut.WithLabelValues("a0", "a1").Inc()
	ticker := time.NewTicker(4 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			sut.WithLabelValues("a0", "a1").Inc()
		}
	}()
	defer ticker.Stop()

	// send metric once
	sut.WithLabelValues("b0", "b1").Inc()

	// then
	requireDataSeries(t, registry, 2)
	require.Len(t, sut.cache, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then
	requireDataSeries(t, registry, 1)
	require.Len(t, sut.cache, 1)

	// and send metric once
	sut.WithLabelValues("b0", "b1").Inc()

	// then check there are 2 data series
	requireDataSeries(t, registry, 2)
	require.Len(t, sut.cache, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then check there is 1 data series
	requireDataSeries(t, registry, 1)
	require.Len(t, sut.cache, 1)
}

func TestPrometheusCounterVecWithoutTtl(t *testing.T) {
	t.Parallel()

	// given
	registry := prometheus.NewRegistry()

	labels := []string{"flag", "family"}

	sut := NewCounterVec(prometheus.CounterOpts{
		Namespace: "ns",
		Name:      "metric_name",
	}, labels, 0)

	registry.MustRegister(sut)

	// when
	sut.WithLabelValues("a0", "a1").Inc()
	sut.WithLabelValues("b0", "b1").Inc()

	// then
	assert.Equal(t, time.Duration(0), sut.ttl)
	require.Len(t, sut.cache, 0)

	requireDataSeries(t, registry, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then
	require.Len(t, sut.cache, 0)
	requireDataSeries(t, registry, 2)
}

func TestPrometheusHistogramVecWithTtl(t *testing.T) {
	t.Parallel()

	// given
	registry := prometheus.NewRegistry()

	labels := []string{"flag", "family"}

	sut := NewTTLHistogramVecWithReconciliation(prometheus.HistogramOpts{
		Namespace: "ns",
		Name:      "metric_name",
	}, labels, 10*time.Millisecond, 10*time.Millisecond)
	defer sut.gcTicker.Stop()

	registry.MustRegister(sut)

	// when
	// send metrics each millisecond
	sut.WithLabelValues("a0", "a1").Observe(0.1)
	ticker := time.NewTicker(4 * time.Millisecond)
	go func() {
		for _ = range ticker.C {
			sut.WithLabelValues("a0", "a1").Observe(0.1)
		}
	}()
	defer ticker.Stop()

	// send metric once
	sut.WithLabelValues("b0", "b1").Observe(0.2)

	// then
	requireDataSeries(t, registry, 2)
	require.Len(t, sut.cache, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then
	requireDataSeries(t, registry, 1)
	require.Len(t, sut.cache, 1)

	// and send metric once
	sut.WithLabelValues("b0", "b1").Observe(0.2)

	// then check there are 2 data series
	requireDataSeries(t, registry, 2)
	require.Len(t, sut.cache, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then check there is 1 data series
	requireDataSeries(t, registry, 1)
	require.Len(t, sut.cache, 1)
}

func TestPrometheusHistogramVecWithoutTtl(t *testing.T) {
	t.Parallel()

	// given
	registry := prometheus.NewRegistry()

	labels := []string{"flag", "family"}

	sut := NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "ns",
		Name:      "metric_name",
	}, labels, 0)

	registry.MustRegister(sut)

	// when
	sut.WithLabelValues("a0", "a1").Observe(0.1)
	sut.WithLabelValues("b0", "b1").Observe(0.2)

	// then
	assert.Equal(t, time.Duration(0), sut.ttl)
	require.Len(t, sut.cache, 0)

	requireDataSeries(t, registry, 2)

	// and let time pass
	time.Sleep(25 * time.Millisecond)

	// then
	require.Len(t, sut.cache, 0)
	requireDataSeries(t, registry, 2)
}

func requireDataSeries(t *testing.T, registry *prometheus.Registry, series int) {
	metricFamilies, err := registry.Gather()
	require.NoError(t, err)

	assert.Equal(t, "ns_metric_name", *metricFamilies[0].Name)
	require.Len(t, metricFamilies[0].Metric, series)
}
