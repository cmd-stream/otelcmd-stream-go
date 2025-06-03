package semconv

import (
	"net"

	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/sender-go/hooks"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

func NewCmdStreamClient[T any](remoteAddr net.Addr,
	meter metric.Meter) (client CmdStreamClient[T]) {
	var (
		fn = func(meter metric.Meter) (cmdCounter metric.Int64Counter,
			resultCounter metric.Int64Counter,
			cmdSizeHistogram metric.Int64Histogram,
			resultSizeHistogram metric.Int64Histogram,
			cmdDurationHistogram metric.Float64Histogram,
			resultDurationHistogram metric.Float64Histogram,
		) {
			cmdCounter, err := meter.Int64Counter(
				semconv.CmdStreamClientCommandCountName,
				metric.WithUnit(semconv.CmdStreamClientCommandCountUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandCountDescription),
			)
			handleErr(err)

			resultCounter, err = meter.Int64Counter(
				semconv.CmdStreamClientResultCountName,
				metric.WithUnit(semconv.CmdStreamClientResultCountUnit),
				metric.WithDescription(semconv.CmdStreamClientResultCountDescription),
			)
			handleErr(err)

			cmdSizeHistogram, err = meter.Int64Histogram(
				semconv.CmdStreamClientCommandSizeName,
				metric.WithUnit(semconv.CmdStreamClientCommandSizeUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandSizeDescription),
			)
			handleErr(err)

			resultSizeHistogram, err = meter.Int64Histogram(
				semconv.CmdStreamClientResultSizeName,
				metric.WithUnit(semconv.CmdStreamClientResultSizeUnit),
				metric.WithDescription(semconv.CmdStreamClientResultSizeDescription),
			)
			handleErr(err)

			cmdDurationHistogram, err = meter.Float64Histogram(
				semconv.CmdStreamClientCommandDurationName,
				metric.WithUnit(semconv.CmdStreamClientCommandDurationUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandDurationDescription),
			)
			handleErr(err)

			resultDurationHistogram, err = meter.Float64Histogram(
				semconv.CmdStreamClientResultDurationName,
				metric.WithUnit(semconv.CmdStreamClientResultDurationUnit),
				metric.WithDescription(semconv.CmdStreamClientResultDurationDescription),
			)
			handleErr(err)
			return
		}
		common = NewCmdStreamCommon[T](meter, fn)
	)
	return CmdStreamClient[T]{common, AddrAttrs(remoteAddr)}
}

type CmdStreamClient[T any] struct {
	CmdStreamCommon[T]
	addrAttrs []attribute.KeyValue
}

func (c CmdStreamClient[T]) SpanAttrs(sentCmd hooks.SentCmd[T],
	addAttrs []attribute.KeyValue) (attrs []attribute.KeyValue) {
	/*
		net.peer.address
		net.peer.port
		network.protocol.name
		cmd-stream.command.seq
		cmd-stream.command.size
	*/
	var (
		l1 = len(addAttrs)
		l2 = len(c.addrAttrs)
		l  = l1 + l2
	)
	attrs = make([]attribute.KeyValue, l, l+2)
	copy(attrs, addAttrs)
	copy(attrs[l1:], c.addrAttrs)
	attrs = append(attrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
	attrs = append(attrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))
	return
}
