package middleware

import (
	"context"
	"time"

	"github.com/Andronzi/credit-origination/pkg/logger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func InitTracer() (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint("host.docker.internal:4317"),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
	)

	otel.SetTracerProvider(tp)
	return tp, nil
}

func TracingInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	ctx = extractTraceContext(ctx)

	ctx, span := otel.Tracer("grpc").Start(ctx, info.FullMethod,
		trace.WithAttributes(
			semconv.RPCSystemKey.String("grpc"),
			semconv.RPCMethodKey.String(info.FullMethod),
		))
	defer span.End()

	startTime := time.Now()
	logger.Logger.Debug("Starting request",
		zap.String("method", info.FullMethod),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
	)

	res, err := handler(ctx, req)
	duration := time.Since(startTime)

	statusCode := codes.OK
	if err != nil {
		if s, ok := status.FromError(err); ok {
			statusCode = s.Code()
		}
		span.RecordError(err)
	}

	logger.Logger.Info("Request completed",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", duration),
		zap.String("trace_id", span.SpanContext().TraceID().String()),
		zap.String("span_id", span.SpanContext().SpanID().String()),
		zap.String("status", statusCode.String()),
		zap.Error(err),
	)

	return res, err
}

func extractTraceContext(ctx context.Context) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	traceParent := md.Get("traceparent")
	if len(traceParent) > 0 {
		ctx = metadata.NewOutgoingContext(ctx, metadata.Pairs("traceparent", traceParent[0]))
	}

	return otel.GetTextMapPropagator().Extract(
		ctx,
		propagation.HeaderCarrier(md),
	)
}
