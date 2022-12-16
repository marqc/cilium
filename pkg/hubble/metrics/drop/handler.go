// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package drop

import (
	"context"

	"github.com/cilium/cilium/pkg/hubble/metrics/util"
	"github.com/prometheus/client_golang/prometheus"

	flowpb "github.com/cilium/cilium/api/v1/flow"
	v1 "github.com/cilium/cilium/pkg/hubble/api/v1"
	"github.com/cilium/cilium/pkg/hubble/metrics/api"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"
)

type dropHandler struct {
	drops   *util.CounterVec
	context *api.ContextOptions
}

func (d *dropHandler) Init(registry *prometheus.Registry, options api.Options) error {
	c, err := api.ParseContextOptions(options)
	if err != nil {
		return err
	}
	d.context = c

	contextLabels := d.context.GetLabelNames()
	labels := append(contextLabels, "reason", "protocol")

	d.drops = util.NewCounterVec(prometheus.CounterOpts{
		Namespace: api.DefaultPrometheusNamespace,
		Name:      "drop_total",
		Help:      "Number of drops",
	}, labels, c.TTL)

	registry.MustRegister(d.drops)
	return nil
}

func (d *dropHandler) Status() string {
	return d.context.Status()
}

func (d *dropHandler) ProcessFlow(ctx context.Context, flow *flowpb.Flow) error {
	if flow.GetVerdict() != flowpb.Verdict_DROPPED {
		return nil
	}

	contextLabels, err := d.context.GetLabelValues(flow)
	if err != nil {
		return err
	}

	labels := append(contextLabels, monitorAPI.DropReason(uint8(flow.GetDropReason())), v1.FlowProtocol(flow))

	d.drops.WithLabelValues(labels...).Inc()
	return nil
}
