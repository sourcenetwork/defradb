// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

//go:build telemetry

package telemetry

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	_ Tracer = (*otelTracer)(nil)
	_ Span   = (*otelSpan)(nil)
)

type otelTracer struct {
	inner trace.Tracer
}

func NewTracer() Tracer {
	name, _ := callerInfo(2)
	tracer := otel.Tracer(name)
	return &otelTracer{tracer}
}

func (t otelTracer) Start(ctx context.Context) (context.Context, Span) {
	_, name := callerInfo(2)
	ctx, span := t.inner.Start(ctx, name)
	return ctx, &otelSpan{span}
}

type otelSpan struct {
	inner trace.Span
}

func (s *otelSpan) End() {
	s.inner.End()
}

// ConfigureTelemetry configures the global telemetry providers for
// defradb and any dependencies that use the OpenTelemetry SDK.
func ConfigureTelemetry(ctx context.Context, version string) error {
	opts := []resource.Option{
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("DefraDB"),
			semconv.ServiceVersionKey.String(version),
		),
		// include all OS info
		resource.WithOS(),
		// include all process info
		resource.WithProcess(),
	}
	res, err := resource.New(ctx, opts...)
	if err != nil {
		return err
	}
	// default to http exporter for traces
	spanExporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return err
	}
	// default to http exporter for metrics
	metricExporter, err := otlpmetrichttp.New(ctx)
	if err != nil {
		return err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(spanExporter),
	)
	// runtime metrics adds info from the Go runtime
	// for more info see the link below:
	// https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/runtime
	runtimeReader := sdkmetric.NewPeriodicReader(
		metricExporter,
		sdkmetric.WithProducer(runtime.NewProducer()),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(runtimeReader),
	)
	// set the global meter provider for all otel instances
	otel.SetMeterProvider(meterProvider)
	// set the global trace provider for all otel instances
	otel.SetTracerProvider(tracerProvider)
	return nil
}
