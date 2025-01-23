package cli

import (
	"context"

	"github.com/sourcenetwork/defradb/version"
	"github.com/spf13/viper"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// configureTelemetry configures the global telemetry providers for
// defradb and any dependencies that use the OpenTelemetry SDK.
func configureTelemetry(ctx context.Context, cfg *viper.Viper) error {
	if !cfg.GetBool("telemetry.enabled") {
		return nil
	}
	ver, err := version.NewDefraVersion()
	if err != nil {
		return err
	}
	opts := []resource.Option{
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceNameKey.String("DefraDB"),
			semconv.ServiceVersionKey.String(ver.Release),
		),
		resource.WithOS(),
		resource.WithProcess(),
	}
	res, err := resource.New(ctx, opts...)
	if err != nil {
		return err
	}
	spanOptions := []otlptracehttp.Option{
		otlptracehttp.WithEndpointURL(cfg.GetString("telemetry.endpoint")),
	}
	spanExporter, err := otlptracehttp.New(ctx, spanOptions...)
	if err != nil {
		return err
	}
	metricOptions := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpointURL(cfg.GetString("telemetry.endpoint")),
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricOptions...)
	if err != nil {
		return err
	}
	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithBatcher(spanExporter),
	)
	runtimeReader := sdkmetric.NewPeriodicReader(
		metricExporter,
		sdkmetric.WithProducer(runtime.NewProducer()),
	)
	meterProvider := sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(runtimeReader),
	)
	otel.SetMeterProvider(meterProvider)
	otel.SetTracerProvider(tracerProvider)
	return nil
}
