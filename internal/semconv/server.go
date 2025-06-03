package semconv

import (
	"net"

	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func NewCmdStreamServer[T any](localAddr net.Addr,
	meter metric.Meter) (client CmdStreamServer[T]) {
	var (
		fn = func(meter metric.Meter) (cmdCounter metric.Int64Counter,
			resultCounter metric.Int64Counter,
			cmdSizeHistogram metric.Int64Histogram,
			resultSizeHistogram metric.Int64Histogram,
			cmdDurationHistogram metric.Float64Histogram,
			resultDurationHistogram metric.Float64Histogram,
		) {
			cmdCounter, err := meter.Int64Counter(
				semconv.CmdStreamServerCommandCountName,
				metric.WithUnit(semconv.CmdStreamServerCommandCountUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandCountDescription),
			)
			handleErr(err)

			cmdSizeHistogram, err = meter.Int64Histogram(
				semconv.CmdStreamServerCommandSizeName,
				metric.WithUnit(semconv.CmdStreamServerCommandSizeUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandSizeDescription),
			)
			handleErr(err)

			cmdDurationHistogram, err = meter.Float64Histogram(
				semconv.CmdStreamServerCommandDurationName,
				metric.WithUnit(semconv.CmdStreamServerCommandDurationUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandDurationDescription),
			)
			handleErr(err)

			resultCounter, err = meter.Int64Counter(
				semconv.CmdStreamServerResultCountName,
				metric.WithUnit(semconv.CmdStreamServerResultCountUnit),
				metric.WithDescription(semconv.CmdStreamServerResultCountDescription),
			)
			handleErr(err)

			resultSizeHistogram, err = meter.Int64Histogram(
				semconv.CmdStreamServerResultSizeName,
				metric.WithUnit(semconv.CmdStreamServerResultSizeUnit),
				metric.WithDescription(semconv.CmdStreamServerResultSizeDescription),
			)
			handleErr(err)

			resultDurationHistogram, err = meter.Float64Histogram(
				semconv.CmdStreamServerResultDurationName,
				metric.WithUnit(semconv.CmdStreamServerResultDurationUnit),
				metric.WithDescription(semconv.CmdStreamServerResultDurationDescription),
			)
			handleErr(err)
			return
		}
		common = NewCmdStreamCommon[T](meter, fn)
	)
	return CmdStreamServer[T]{common}
}

type CmdStreamServer[T any] struct {
	CmdStreamCommon[T]
}

func (c CmdStreamServer[T]) SpanAttrs(remoteAddr net.Addr,
	addAttrs []attribute.KeyValue) (attrs []attribute.KeyValue) {
	/*
		net.peer.ip
		net.peer.port
		network.protocol.name
	*/
	var (
		addrAttrs = AddrAttrs(remoteAddr)
		l1        = len(addAttrs)
		l2        = len(addrAttrs)
	)
	attrs = make([]attribute.KeyValue, l1+l2)
	copy(attrs, addAttrs)
	copy(attrs[l1:], addrAttrs)
	return
}

func (c CmdStreamServer[T]) SpanResultEventAttrs(sentCmd hooks.SentCmd[T],
	recvResult hooks.ReceivedResult,
	addAttrs []attribute.KeyValue,
) (attrs []attribute.KeyValue) {
	/*
		cmd-stream.result.seq
		cmd-stream.result.size
		cmd-stream.result.type
	*/
	l := len(addAttrs)
	attrs = make([]attribute.KeyValue, l, l+3)
	copy(attrs, addAttrs)
	attrs = append(attrs, semconv.CmdStreamResultSeqKey.Int64(int64(recvResult.Seq)))
	attrs = append(attrs, semconv.CmdStreamResultSizeKey.Int64(int64(recvResult.Size)))
	attrs = append(attrs, c.ResultTypeAttr(recvResult.Result))
	return
}
