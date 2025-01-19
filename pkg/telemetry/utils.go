package telemetry

import (
	"context"
	"runtime"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

func StartSpanWithFuncName(ctx context.Context) (context.Context, trace.Span) {
	spanName := "unknown"

	pc, _, _, ok := runtime.Caller(1)
	if ok {
		callerFuncName := runtime.FuncForPC(pc).Name()
		if callerFuncName != "" {
			spanName = callerFuncName
		}
	}

	return Tracer.Start(ctx, spanName)
}

func GetMapCarrierFromContext(ctx context.Context) map[string]string {
	telemetryMeta := map[string]string{}
	otel.GetTextMapPropagator().Inject(ctx, propagation.MapCarrier(telemetryMeta))

	return telemetryMeta
}

func RestoreContextFromMapCarrier(ctx context.Context, telemetryMeta map[string]string) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(telemetryMeta))
}
