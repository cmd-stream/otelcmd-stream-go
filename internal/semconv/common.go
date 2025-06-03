package semconv

import (
	"context"
	"net"
	"reflect"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	otel_semconv "go.opentelemetry.io/otel/semconv/v1.30.0"
)

type typed interface {
	TypeStr() string
}

type unitsFn func(meter metric.Meter) (cmdCounter metric.Int64Counter,
	resultCounter metric.Int64Counter,
	cmdSizeHistogram metric.Int64Histogram,
	resultSizeHistogram metric.Int64Histogram,
	cmdDurationHistogram metric.Float64Histogram,
	resultDurationHistogram metric.Float64Histogram,
)

func NewCmdStreamCommon[T any](meter metric.Meter, fn unitsFn) (
	c CmdStreamCommon[T]) {
	if meter == nil {
		c.cmdCounter = noop.Int64Counter{}
		c.resultCounter = noop.Int64Counter{}
		c.cmdSizeHistogram = noop.Int64Histogram{}
		c.resultSizeHistogram = noop.Int64Histogram{}
		c.cmdDurationHistogram = noop.Float64Histogram{}
		c.resultDurationHistogram = noop.Float64Histogram{}
		return
	}
	c.cmdCounter, c.resultCounter, c.cmdSizeHistogram, c.resultSizeHistogram,
		c.cmdDurationHistogram, c.resultDurationHistogram = fn(meter)
	return
}

type CmdStreamCommon[T any] struct {
	cmdCounter    metric.Int64Counter
	resultCounter metric.Int64Counter

	cmdSizeHistogram    metric.Int64Histogram
	resultSizeHistogram metric.Int64Histogram

	cmdDurationHistogram    metric.Float64Histogram
	resultDurationHistogram metric.Float64Histogram
}

func (c CmdStreamCommon[T]) RecordCmdMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	status semconv.CmdStreamCommandStatus,
	elapsedTime float64,
	addAttrs []attribute.KeyValue,
) {
	op := c.cmdMetricOption(sentCmd.Cmd, status, addAttrs)
	c.cmdCounter.Add(ctx, 1, op)
	c.cmdSizeHistogram.Record(ctx, int64(sentCmd.Size), op)
	c.cmdDurationHistogram.Record(ctx, elapsedTime, op)
}

func (c CmdStreamCommon[T]) RecordResultMetrics(ctx context.Context,
	sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	elapsedTime float64,
	addAttrs []attribute.KeyValue,
) {
	op := c.resultMetricsOption(sentCmd.Cmd, recvResult.Result, addAttrs)
	c.resultCounter.Add(ctx, 1, op)
	c.resultSizeHistogram.Record(ctx, int64(recvResult.Size), op)
	c.resultDurationHistogram.Record(ctx, elapsedTime, op)
}

func (c CmdStreamCommon[T]) CmdTypeAttr(cmd core.Cmd[T]) attribute.KeyValue {
	return semconv.CmdStreamCommandTypeKey.String(TypeStr(cmd))
}

func (c CmdStreamCommon[T]) ResultTypeAttr(result core.Result) attribute.KeyValue {
	return semconv.CmdStreamResultTypeKey.String(TypeStr(result))
}

func (c CmdStreamCommon[T]) ErrorTypeAttr(err error) attribute.KeyValue {
	return otel_semconv.ErrorTypeKey.String(TypeStr(err))
}

func (c CmdStreamCommon[T]) TypeAttrValue(t reflect.Type) (value string) {
	return t.String()
}

func (c CmdStreamCommon[T]) cmdMetricOption(cmd core.Cmd[T],
	status semconv.CmdStreamCommandStatus,
	addAttrs []attribute.KeyValue,
) metric.MeasurementOption {
	var (
		l     = len(addAttrs)
		attrs = make([]attribute.KeyValue, l, l+2)
	)
	copy(attrs, addAttrs)
	attrs = append(attrs, c.CmdTypeAttr(cmd))
	attrs = append(attrs, semconv.CmdStreamCommandStatusKey.String(string(status)))
	return metric.WithAttributeSet(attribute.NewSet(attrs...))
}

func (c CmdStreamCommon[T]) resultMetricsOption(cmd core.Cmd[T],
	result core.Result, addAttrs []attribute.KeyValue) metric.MeasurementOption {
	var (
		l     = len(addAttrs)
		attrs = make([]attribute.KeyValue, l, l+2)
	)
	copy(attrs, addAttrs)
	attrs = append(attrs, c.CmdTypeAttr(cmd))
	attrs = append(attrs, c.ResultTypeAttr(result))
	return metric.WithAttributeSet(attribute.NewSet(attrs...))
}

func TypeStr(a any) (str string) {
	if typedCmd, ok := a.(typed); ok {
		return typedCmd.TypeStr()
	}
	return reflect.TypeOf(a).Name()
}

func AddrAttrs(addr net.Addr) []attribute.KeyValue {
	var (
		address      = "undefined"
		port         = 0
		protocolName = "undefined"
	)
	if addr != nil {
		address, port = addressPort(addr)
		protocolName = protoName(addr)
	}
	return []attribute.KeyValue{
		otel_semconv.NetworkPeerAddress(address),
		otel_semconv.NetworkPeerPort(port),
		otel_semconv.NetworkProtocolName(protocolName),
	}
}
