package otelcs

import (
	"context"
	"time"

	"github.com/cmd-stream/core-go"
	internal_semconv "github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// NewHooksFactory creates a new HooksFactory.
func NewHooksFactory[T any](ops ...SetOption[T]) HooksFactory[T] {
	o := Options[T]{
		SpanStartOptions:  []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindClient)},
		SpanNameFormatter: defaultClientSpanNameFormatter[T],
		Propagator:        otel.GetTextMapPropagator(),
		// Look BeforeSend method.
		// TracerProvider:    otel.GetTracerProvider(),
		MeterProvider: otel.GetMeterProvider(),
	}
	Apply(ops, &o)
	return HooksFactory[T]{o}
}

// HooksFactory is an implementation of the hooks.HooksFactory interface from
// the sender module. It is responsible for creating Hooks instances configured
// with OpenTelemetry tracing and metrics options.
type HooksFactory[T any] struct {
	options Options[T]
}

func (f HooksFactory[T]) New() hooks.Hooks[T] {
	return &Hooks[T]{
		semconv: internal_semconv.NewCmdStreamClient[T](f.options.ServerAddr, f.options.Meter),
		options: f.options,
	}
}

// Hooks is an implementation of the hooks.Hooks interface from the sender
// module. It provides OpenTelemetry-based instrumentation for the cmd-stream
// sender.
type Hooks[T any] struct {
	startTime time.Time
	span      trace.Span
	semconv   internal_semconv.CmdStreamClient[T]
	options   Options[T]
}

func (h *Hooks[T]) BeforeSend(ctx context.Context, cmd core.Cmd[T]) (context.Context, error) {
	h.startTime = time.Now()
	tracer := h.options.Tracer
	if tracer == nil {
		if span := trace.SpanFromContext(ctx); span.SpanContext().IsValid() {
			tracer = newTracer(span.TracerProvider())
		} else {
			tracer = newTracer(otel.GetTracerProvider())
		}
	}
	var actx context.Context
	actx, h.span = tracer.Start(ctx, h.options.SpanNameFormatter(cmd),
		h.options.SpanStartOptions...)

	if tcmd, ok := cmd.(traceCmd[T]); ok {
		carrier := propagation.MapCarrier{}
		h.options.Propagator.Inject(actx, carrier)
		tcmd.SetCarrier(carrier)
	}
	return actx, nil
}

func (h *Hooks[T]) OnError(ctx context.Context, sentCmd hooks.SentCmd[T],
	err error) {
	if errAttr := h.semconv.ErrorTypeAttr(err); errAttr.Valid() {
		h.span.SetAttributes(errAttr)
	}
	h.setSpanAttributes(sentCmd)
	h.span.SetStatus(codes.Error, err.Error())
	h.span.End()
}

func (h *Hooks[T]) OnResult(ctx context.Context, sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult, err error) {
	elapsedTime := ElapsedTime(h.startTime)

	if err != nil {
		h.recordCmdMetrics(ctx, sentCmd, semconv.Failed, elapsedTime)
		h.OnError(ctx, sentCmd, err)
		return
	}

	h.setSpanAttributes(sentCmd)
	h.setSpanResultEventAttributes(sentCmd, recvResult)

	h.recordResultMetrics(ctx, sentCmd, recvResult, elapsedTime)
	if recvResult.Result.LastOne() {
		h.recordCmdMetrics(ctx, sentCmd, semconv.Ok, elapsedTime)
		h.span.End()
	}
}

func (h *Hooks[T]) OnTimeout(ctx context.Context, sentCmd hooks.SentCmd[T],
	err error) {
	elapsedTime := ElapsedTime(h.startTime)

	h.recordCmdMetrics(ctx, sentCmd, semconv.Timeout, elapsedTime)
	h.OnError(ctx, sentCmd, err)
}

func (h *Hooks[T]) setSpanAttributes(sentCmd hooks.SentCmd[T]) {
	var addAttrs []attribute.KeyValue
	if h.options.SpanAttributesFn != nil {
		addAttrs = h.options.SpanAttributesFn(h.options.ServerAddr, sentCmd)
	}
	h.span.SetAttributes(h.semconv.SpanAttrs(sentCmd, addAttrs)...)
}

func (h *Hooks[T]) setSpanResultEventAttributes(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult) {
	if h.options.SpanResultEventAttributesFn != nil {
		h.span.AddEvent(internal_semconv.ResultEventName, trace.WithAttributes(
			h.options.SpanResultEventAttributesFn(sentCmd, recvResult)...,
		))
	}
}

func (h *Hooks[T]) recordCmdMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	status semconv.CmdStreamCommandStatus,
	elapsedTime float64,
) {
	var addAttrs []attribute.KeyValue
	if h.options.CmdMetricAttributesFn != nil {
		addAttrs = h.options.CmdMetricAttributesFn(sentCmd, status, elapsedTime)
	}
	h.semconv.RecordCmdMetrics(ctx, sentCmd, status, elapsedTime, addAttrs)
}

func (h *Hooks[T]) recordResultMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	elapsedTime float64,
) {
	var addAttrs []attribute.KeyValue
	if h.options.ResultMetricAttributesFn != nil {
		addAttrs = h.options.ResultMetricAttributesFn(sentCmd, recvResult, elapsedTime)
	}
	h.semconv.RecordResultMetrics(ctx, sentCmd, recvResult, elapsedTime, addAttrs)
}

func newTracer(tp trace.TracerProvider) trace.Tracer {
	return tp.Tracer(ScopeName, trace.WithInstrumentationVersion(Version()))
}
