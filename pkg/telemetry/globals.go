package telemetry

import (
	metricnoop "go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

// Each of these variables is overwritten by telemetry.Setup.
var (
	Tracer = tracenoop.NewTracerProvider().Tracer("noop")
	Meter  = metricnoop.NewMeterProvider().Meter("noop")
)
