package telemetry

import (
	"context"
	"errors"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type TelemetryConfig struct {
	ServiceName        string
	TracerAndMeterName string
}

func Setup(ctx context.Context, config *TelemetryConfig) (func(context.Context) error, error) {
	var shutdownFuncs []func(context.Context) error
	shutdown := func(ctx context.Context) error {
		var err error
		for _, fn := range shutdownFuncs {
			err = errors.Join(err, fn(ctx))
		}

		return err
	}

	handleErrInSetup := func(err error) error {
		return errors.Join(err, shutdown(ctx))
	}

	res, err := resource.New(
		ctx,
		resource.WithOS(),
		resource.WithProcess(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithTelemetrySDK(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(config.ServiceName),
		),
	)
	if err != nil {
		return nil, handleErrInSetup(err)
	}

	propagator := propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	)
	otel.SetTextMapPropagator(propagator)

	tracerProvider, err := newTraceProvider(ctx, res)
	if err != nil {
		return nil, handleErrInSetup(err)
	}
	shutdownFuncs = append(shutdownFuncs, tracerProvider.Shutdown)
	otel.SetTracerProvider(tracerProvider)
	Tracer = tracerProvider.Tracer(config.TracerAndMeterName)

	meterProvider, err := newMeterProvider(ctx, res)
	if err != nil {
		return nil, handleErrInSetup(err)
	}
	shutdownFuncs = append(shutdownFuncs, meterProvider.Shutdown)
	otel.SetMeterProvider(meterProvider)
	Meter = meterProvider.Meter(config.TracerAndMeterName)

	// Start go runtime metric collection.
	// See: https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/runtime
	err = runtime.Start(runtime.WithMinimumReadMemStatsInterval(time.Second))
	if err != nil {
		return nil, handleErrInSetup(err)
	}

	return shutdown, nil
}

func newTraceProvider(ctx context.Context, res *resource.Resource) (*trace.TracerProvider, error) {
	// The following configuration is not working as expected.
	// See: https://github.com/open-telemetry/opentelemetry-go/issues/2940
	// traceExporter, err := otlptracegrpc.New(
	// 	ctx,
	// 	otlptracegrpc.WithEndpoint("localhost:5003"),
	// 	otlptracegrpc.WithDialOption(grpc.WithTransportCredentials(insecure.NewCredentials())),
	// )

	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint("localhost:4317"),
	)
	traceExporter, err := otlptrace.New(ctx, traceClient)
	if err != nil {
		return nil, err
	}

	traceProvider := trace.NewTracerProvider(
		// Set to 1s for demonstrative purposes.
		// In production, it is recommended to use a longer interval fitting monitoring needs.
		trace.WithBatcher(traceExporter, trace.WithBatchTimeout(1*time.Second)),
		trace.WithResource(res),
	)

	return traceProvider, nil
}

func newMeterProvider(ctx context.Context, res *resource.Resource) (*metric.MeterProvider, error) {
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithInsecure(),
		otlpmetricgrpc.WithEndpoint("localhost:4317"),
	)
	if err != nil {
		return nil, err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithReader(
			metric.NewPeriodicReader(
				metricExporter,
				// Set to 5s for demonstrative purposes.
				// In production, it is recommended to use a longer interval fitting monitoring needs.
				metric.WithInterval(5*time.Second),
				// runtime.NewProducer() collects go scheduler metrics.
				metric.WithProducer(runtime.NewProducer()),
			),
		),
		metric.WithResource(res),
	)

	return meterProvider, nil
}
