// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package metric

import (
	"context"
	"encoding/json"

	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncint64"
	"go.opentelemetry.io/otel/metric/unit"
	otelMetric "go.opentelemetry.io/otel/sdk/metric"
)

var _ Metric = (*Meter)(nil)

// Metric interface attempts to abstract high-level aspects of the observability features,
// this would make it much easier to swap the base sdk without having to change too
// many implementation details.
type Metric interface {
	// Register gives a name to the metric and initializes the provider.
	Register(name string)

	// Dump is responsible to read the metrics and output all the gathered data.
	Dump(ctx context.Context) (any, error)

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

func NewMeter() Meter {
	return Meter{}
}

func (m *Meter) Register(name string) {
	m.provider = m.newManualProvider()
	m.meter = m.provider.Meter(name)
}

func (m *Meter) Dump(ctx context.Context) (any, error) {
	return m.reader.Collect(ctx)
}

func (m *Meter) Close(ctx context.Context) error {
	return m.provider.Shutdown(ctx)
}

func (m *Meter) GetSyncHistogram(
	name string,
	unit unit.Unit,
) (syncint64.Histogram, error) {
	return m.meter.SyncInt64().Histogram(
		name,
		instrument.WithUnit(unit),
	)
}

func (m *Meter) GetSyncCounter(
	name string,
	unit unit.Unit,
) (syncint64.Counter, error) {
	return m.meter.SyncInt64().Counter(
		name,
		instrument.WithUnit(unit),
	)
}

func (m *Meter) DumpScopeMetricsString(ctx context.Context) (string, error) {
	data, err := m.reader.Collect(ctx)
	if err != nil {
		return "", err
	}
	jsonBytes, err := json.MarshalIndent(data.ScopeMetrics, "", "  ")

	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

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
