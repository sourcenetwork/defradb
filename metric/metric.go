// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

/*
Package metric provides the observability system.

It is a wrapper around the opentelemetry metric package.
*/
package metric

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/metric"
	otelMetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
)

var _ Metric = (*Meter)(nil)

// Metric interface attempts to abstract high-level aspects of the observability features,
// this would make it much easier to swap the base sdk without having to change too
// many implementation details.
type Metric interface {
	// Register gives a name to the metric and initializes the provider.
	Register(name string)

	// Dump is responsible to read the metrics and output all the gathered data.
	Dump(ctx context.Context) (*metricdata.ResourceMetrics, error)

	// Close shutsdown the meter.
	Close(ctx context.Context) error
}

// Meter is currently wrapping opentelemetry meter functionalities while adhering to
// the Metric interface functionality.
type Meter struct {
	reader   otelMetric.Reader
	provider *otelMetric.MeterProvider
	meter    metric.Meter
}

// NewMeter returns a new meter.
func NewMeter() Meter {
	return Meter{}
}

// Register gives a name to the metric and initializes the provider.
func (m *Meter) Register(name string) {
	m.provider = m.newManualProvider()
	m.meter = m.provider.Meter(name)
}

// Dump is responsible to read the metrics and output all the gathered data.
func (m *Meter) Dump(ctx context.Context) (*metricdata.ResourceMetrics, error) {
	out := &metricdata.ResourceMetrics{}
	if err := m.reader.Collect(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

// Close shutsdown the meter.
func (m *Meter) Close(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}

// GetSyncHistogram returns a new histogram with the given name and unit.
func (m *Meter) GetSyncHistogram(
	name string,
	unit string,
) (metric.Int64Histogram, error) {
	return m.meter.Int64Histogram(
		name,
		metric.WithUnit(unit),
	)
}

// GetSyncCounter returns a new counter with the given name and unit.
func (m *Meter) GetSyncCounter(
	name string,
	unit string,
) (metric.Int64Counter, error) {
	return m.meter.Int64Counter(
		name,
		metric.WithUnit(unit),
	)
}

// DumpScopeMetricsString returns a string representation of the metrics.
func (m *Meter) DumpScopeMetricsString(ctx context.Context) (string, error) {
	out := &metricdata.ResourceMetrics{}
	if err := m.reader.Collect(ctx, out); err != nil {
		return "", err
	}

	jsonBytes, err := json.MarshalIndent(out.ScopeMetrics, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// Get returns the meter.
func (m *Meter) Get() metric.Meter {
	return m.meter
}

func (m *Meter) newManualProvider() *otelMetric.MeterProvider {
	// Register a manual reader.
	m.reader = otelMetric.NewManualReader()
	return otelMetric.NewMeterProvider(
		otelMetric.WithReader(m.reader),
	)
}
