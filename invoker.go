package otelcs

import (
	"context"
	"net"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/handler-go"
	internal_semconv "github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// NewInvoker creates a new invoker that wraps the specified one, adding
// OpenTelemetry-based instrumentation.
func NewInvoker[T any](invoker handler.Invoker[T], ops ...SetOption[T]) Invoker[T] {
	o := Options[T]{
		SpanStartOptions: []trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer),
		},
		SpanNameFormatter: defaultServerSpanNameFormatter[T],
		Propagator:        otel.GetTextMapPropagator(),
		TracerProvider:    otel.GetTracerProvider(),
		MeterProvider:     otel.GetMeterProvider(),
	}
	Apply(ops, &o)
	return Invoker[T]{
		invoker: invoker,
		semconv: internal_semconv.NewCmdStreamServer[T](o.ServerAddr, o.Meter),
		options: o,
	}
}

// Invoker is an implementation of the handler.Invoker interface from the
// handler module. It adds OpenTelemetry-based instrumentation for command
// handling on the server side.
type Invoker[T any] struct {
	invoker handler.Invoker[T]
	semconv internal_semconv.CmdStreamServer[T]
	options Options[T]
}

func (i Invoker[T]) Invoke(ctx context.Context, seq core.Seq, at time.Time,
	bytesRead int, cmd core.Cmd[T], proxy core.Proxy) (err error) {
	startTime := time.Now()

	if tcmd, ok := cmd.(traceCmd[T]); ok {
		ctx = i.options.Propagator.Extract(ctx, propagation.MapCarrier(tcmd.Carrier()))
	}
	// TODO
	// if startTime := StartTimeFromContext(ctx); !startTime.IsZero() {
	// 	opts = append(opts, trace.WithTimestamp(startTime))
	// 	requestStartTime = startTime
	// }
	ctx, span := i.options.Tracer.Start(ctx, i.options.SpanNameFormatter(cmd),
		i.options.SpanStartOptions...)
	sentCmd := hooks.SentCmd[T]{Seq: seq, Size: bytesRead, Cmd: cmd}
	i.setSpanAttributes(span, proxy.RemoteAddr(), sentCmd)

	var (
		callback = func(recvResult hooks.ReceivedResult) {
			i.setSpanResultEventAttributes(span, sentCmd, recvResult)
			i.recordResultMetrics(ctx, sentCmd, recvResult, ElapsedTime(startTime))
		}
		proxyWrap = NewProxy[T](proxy, callback)
	)
	err = i.invoker.Invoke(ctx, seq, at, bytesRead, cmd, proxyWrap)

	status := semconv.Ok
	if err != nil {
		status = semconv.Failed
		if errAttr := i.semconv.ErrorTypeAttr(err); errAttr.Valid() {
			span.SetAttributes(errAttr)
		}
		span.SetStatus(codes.Error, err.Error())
	}
	i.recordCmdMetrics(ctx, sentCmd, status, ElapsedTime(startTime))
	span.End()
	return
}

func (i Invoker[T]) setSpanAttributes(span trace.Span, remoteAddr net.Addr,
	sentCmd hooks.SentCmd[T]) {
	var addAttrs []attribute.KeyValue
	if i.options.SpanAttributesFn != nil {
		addAttrs = i.options.SpanAttributesFn(remoteAddr, sentCmd)
	}
	span.SetAttributes(i.semconv.SpanAttrs(remoteAddr, addAttrs)...)
}

func (i Invoker[T]) setSpanResultEventAttributes(span trace.Span,
	sentCmd hooks.SentCmd[T], recvResult hooks.ReceivedResult) {
	var addAttrs []attribute.KeyValue
	if i.options.SpanResultEventAttributesFn != nil {
		addAttrs = i.options.SpanResultEventAttributesFn(sentCmd, recvResult)
	}
	span.AddEvent(internal_semconv.ResultEventName, trace.WithAttributes(
		i.semconv.SpanResultEventAttrs(sentCmd, recvResult, addAttrs)...,
	))

}

func (i Invoker[T]) recordCmdMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	status semconv.CmdStreamCommandStatus,
	elapsedTime float64,
) {
	var addAttrs []attribute.KeyValue
	if i.options.CmdMetricAttributesFn != nil {
		addAttrs = i.options.CmdMetricAttributesFn(sentCmd, status, elapsedTime)
	}
	i.semconv.RecordCmdMetrics(ctx, sentCmd, status, elapsedTime, addAttrs)
}

func (i Invoker[T]) recordResultMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	elapsedTime float64,
) {
	var addAttrs []attribute.KeyValue
	if i.options.ResultMetricAttributesFn != nil {
		addAttrs = i.options.ResultMetricAttributesFn(sentCmd, recvResult, elapsedTime)
	}
	i.semconv.RecordResultMetrics(ctx, sentCmd, recvResult, elapsedTime, addAttrs)
}
