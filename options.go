package otelcmd

import (
	"net"

	"github.com/cmd-stream/core-go"
	internal_semconv "github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// ScopeName is the instrumentation scope name.
const ScopeName = "github.com/cmd-stream/otelcmd-stream-go"

type SpanNameFormatterFn[T any] func(cmd core.Cmd[T]) string

type CmdMetricAttributesFn[T any] func(sentCmd hooks.SentCmd[T],
	status semconv.CmdStreamCommandStatus,
	elapsedTime float64,
) []attribute.KeyValue

type ResultMetricAttributesFn[T any] func(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	elapsedTime float64,
) []attribute.KeyValue

type SpanAttributesFn[T any] func(remoteAddr net.Addr,
	sentCmd hooks.SentCmd[T]) []attribute.KeyValue

type SpanResultEventAttributesFn[T any] func(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult) []attribute.KeyValue

type Options[T any] struct {
	ServerAddr        net.Addr
	Tracer            trace.Tracer
	Meter             metric.Meter
	SpanStartOptions  []trace.SpanStartOption
	SpanNameFormatter SpanNameFormatterFn[T]

	Propagator     propagation.TextMapPropagator
	TracerProvider trace.TracerProvider
	MeterProvider  metric.MeterProvider

	SpanAttributesFn            SpanAttributesFn[T]
	SpanResultEventAttributesFn SpanResultEventAttributesFn[T]

	CmdMetricAttributesFn    CmdMetricAttributesFn[T]
	ResultMetricAttributesFn ResultMetricAttributesFn[T]
}

// SpanAttributes returns span attributes for the given peer address and sent
// Command. It calls the user-provided SpanAttributesFn if set; otherwise, it
// returns nil.
func (o Options[T]) SpanAttributes(peerAddr net.Addr,
	sentCmd hooks.SentCmd[T]) (attrs []attribute.KeyValue) {
	if o.SpanAttributesFn != nil {
		return o.SpanAttributesFn(peerAddr, sentCmd)
	}
	return
}

// SpanResultEventAttributes returns span event attributes for the given sent
// Command and received result. It calls the user-provided
// SpanResultEventAttributesFn if set; otherwise, it returns nil.
func (o Options[T]) SpanResultEventAttributes(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult) (attrs []attribute.KeyValue) {
	if o.SpanResultEventAttributesFn != nil {
		return o.SpanResultEventAttributesFn(sentCmd, recvResult)
	}
	return
}

// CmdMetricAttributes returns metric attributes for the sent Command, given its
// status and elapsed time. It calls the user-provided CmdMetricAttributesFn if
// set; otherwise, returns nil.
func (o Options[T]) CmdMetricAttributes(sentCmd hooks.SentCmd[T],
	status semconv.CmdStreamCommandStatus,
	elapsedTime float64,
) (attrs []attribute.KeyValue) {
	if o.CmdMetricAttributesFn != nil {
		return o.CmdMetricAttributesFn(sentCmd, status, elapsedTime)
	}
	return
}

// ResultMetricAttributes returns metric attributes for the received Result,
// given the corresponding sent Command and elapsed time. It calls the
// user-provided ResultMetricAttributesFn if set; otherwise, returns nil.
func (o Options[T]) ResultMetricAttributes(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	elapsedTime float64,
) (attrs []attribute.KeyValue) {
	if o.ResultMetricAttributesFn != nil {
		return o.ResultMetricAttributesFn(sentCmd, recvResult, elapsedTime)
	}
	return
}

type SetOption[T any] func(o *Options[T])

// WithServerAddr sets the server address.
func WithServerAddr[T any](addr net.Addr) SetOption[T] {
	return func(o *Options[T]) {
		o.ServerAddr = addr
	}
}

// WithPropagator sets the OpenTelemetry TextMapPropagator.
func WithPropagator[T any](p propagation.TextMapPropagator) SetOption[T] {
	return func(o *Options[T]) {
		o.Propagator = p
	}
}

// WithTracerProvider sets the OpenTelemetry TracerProvider.
func WithTracerProvider[T any](tp trace.TracerProvider) SetOption[T] {
	return func(o *Options[T]) {
		o.TracerProvider = tp
	}
}

// WithMeterProvider sets the OpenTelemetry MeterProvider.
func WithMeterProvider[T any](mp metric.MeterProvider) SetOption[T] {
	return func(o *Options[T]) {
		o.MeterProvider = mp
	}
}

// WithSpanNameFormatter sets the function used to format span names.
func WithSpanNameFormatter[T any](f SpanNameFormatterFn[T]) SetOption[T] {
	return func(o *Options[T]) {
		o.SpanNameFormatter = f
	}
}

// WithSpanStartOption appends a SpanStartOption to be applied when starting
// spans.
func WithSpanStartOption[T any](s trace.SpanStartOption) SetOption[T] {
	return func(o *Options[T]) {
		o.SpanStartOptions = append(o.SpanStartOptions, s)
	}
}

// WithSpanAttributesFn sets the function that returns additional span
// attributes.
func WithSpanAttributesFn[T any](fn SpanAttributesFn[T]) SetOption[T] {
	return func(o *Options[T]) {
		o.SpanAttributesFn = fn
	}
}

// WithSpanResultEventAttributesFn sets the function that returns additional
// span event attributes.
func WithSpanResultEventAttributesFn[T any](
	fn SpanResultEventAttributesFn[T]) SetOption[T] {
	return func(o *Options[T]) {
		o.SpanResultEventAttributesFn = fn
	}
}

// WithCmdMetricAttributesFn sets the function that returns Command metric
// attributes.
func WithCmdMetricAttributesFn[T any](fn CmdMetricAttributesFn[T]) SetOption[T] {
	return func(o *Options[T]) {
		o.CmdMetricAttributesFn = fn
	}
}

// WithResultMetricAttributesFn sets the function that returns Result metric
// attributes.
func WithResultMetricAttributesFn[T any](fn ResultMetricAttributesFn[T]) SetOption[T] {
	return func(o *Options[T]) {
		o.ResultMetricAttributesFn = fn
	}
}

func Apply[T any](ops []SetOption[T], o *Options[T]) {
	for i := range ops {
		if ops[i] != nil {
			ops[i](o)
		}
	}
	if o.TracerProvider != nil {
		o.Tracer = newTracer(o.TracerProvider)
	}
	if o.MeterProvider != nil {
		o.Meter = o.MeterProvider.Meter(
			ScopeName,
			metric.WithInstrumentationVersion(Version()),
		)
	}
}

func defaultClientSpanNameFormatter[T any](cmd core.Cmd[T]) string {
	return "Send " + internal_semconv.TypeStr(cmd)
}

func defaultServerSpanNameFormatter[T any](cmd core.Cmd[T]) string {
	return "Invoke " + internal_semconv.TypeStr(cmd)
}
