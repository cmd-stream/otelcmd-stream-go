package otelcmd

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	bmock "github.com/cmd-stream/core-go/testdata/mock"
	hmock "github.com/cmd-stream/handler-go/testdata/mock"
	internal_semconv "github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/testdata/mock"
	"github.com/cmd-stream/sender-go/hooks"
	asserterror "github.com/ymz-ncnk/assert/error"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	otel_semconv "go.opentelemetry.io/otel/semconv/v1.22.0"
	"go.opentelemetry.io/otel/trace"
	tracenop "go.opentelemetry.io/otel/trace/noop"
)

const (
	CmdSeq  = 1
	CmdSize = 10

	ResultSeq   = 1
	ResultSize  = 20
	Traceparent = "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"
)

type metricVars struct {
	cmdInt64Counter        mock.Int64Counter
	cmdInt64Histogram      mock.Int64Histogram
	cmdFloat64Histogram    mock.Float64Histogram
	resultInt64Counter     mock.Int64Counter
	resultInt64Histogram   mock.Int64Histogram
	resultFloat64Histogram mock.Float64Histogram
}

func TestInvoker(t *testing.T) {

	t.Run("Should work with TraceCmd", func(t *testing.T) {
		otel.SetTracerProvider(tracenop.NewTracerProvider())
		otel.SetMeterProvider(noop.NewMeterProvider())
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

		var (
			wantAddr = &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8080,
			}
			meterProvider  = mock.NewMeterProvider()
			tracerProvider = mock.NewTracerProvider()
			propagator     = propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)
			result = bmock.NewResult()
			cmd    = bmock.NewCmd().RegisterExec(
				func(ctx context.Context, seq core.Seq, at time.Time, receiver any,
					proxy core.Proxy) (err error) {
					_, err = proxy.Send(0, result)
					return
				},
			)
			traceCmd = TraceCmd[any, core.Cmd[any]]{
				MapCarrier: &map[string]string{
					"traceparent": Traceparent,
					// "tracestate":  "vendor1=abc,vendor2=def",
				},
				Cmd: cmd,
			}
			wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd) + " (trace)"
			ops          []SetOption[any]
		)
		otel.SetMeterProvider(meterProvider)
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagator)

		want := newWantVals(wantAddr, wantSpanName, traceCmd, result, semconv.Ok, nil,
			nil, nil, nil, nil, true)
		testInvoke(want, meterProvider, tracerProvider, traceCmd, result, ops, t)
	})

	t.Run("Should work with a regular Cmd", func(t *testing.T) {
		var (
			wantAddr = &net.TCPAddr{
				IP:   net.ParseIP("127.0.0.1"),
				Port: 8080,
			}
			meterProvider  = mock.NewMeterProvider()
			tracerProvider = mock.NewTracerProvider()
			propagator     = propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)
			result = bmock.NewResult()
			cmd    = bmock.NewCmd().RegisterExec(
				func(ctx context.Context, seq core.Seq, at time.Time, receiver any,
					proxy core.Proxy) (err error) {
					_, err = proxy.Send(0, result)
					return
				},
			)
			wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
			ops          []SetOption[any]
		)
		otel.SetMeterProvider(meterProvider)
		otel.SetTracerProvider(tracerProvider)
		otel.SetTextMapPropagator(propagator)

		want := newWantVals(wantAddr, wantSpanName, cmd, result, semconv.Ok, nil,
			nil, nil, nil, nil, true)
		testInvoke(want, meterProvider, tracerProvider, cmd, result, ops, t)
	})

	t.Run("We should be able to set own SpanNameFormatter, Propagator, TracerProvider and MeterProvider",
		func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}
				propagator = propagation.NewCompositeTextMapPropagator(
					propagation.TraceContext{},
					propagation.Baggage{},
				)
				meterProvider  = mock.NewMeterProvider()
				tracerProvider = mock.NewTracerProvider()
				result         = bmock.NewResult()
				cmd            = bmock.NewCmd().RegisterExec(
					func(ctx context.Context, seq core.Seq, at time.Time, receiver any,
						proxy core.Proxy) (err error) {
						_, err = proxy.Send(0, result)
						return
					},
				)
				wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
				ops          = []SetOption[any]{
					WithPropagator[any](propagator),
					WithTracerProvider[any](tracerProvider),
					WithMeterProvider[any](meterProvider),
					WithSpanNameFormatter(func(cmd core.Cmd[any]) string { return wantSpanName }),
				}
			)

			want := newWantVals(wantAddr, wantSpanName, cmd, result, semconv.Ok, nil,
				nil, nil, nil, nil, true)
			testInvoke(want, meterProvider, tracerProvider, cmd, result, ops, t)
		})

	t.Run("We should be able to set SpanStartOptions, and add span/metric attributes",
		func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}
				meterProvider  = mock.NewMeterProvider()
				tracerProvider = mock.NewTracerProvider()
				propagator     = propagation.NewCompositeTextMapPropagator(
					propagation.TraceContext{},
					propagation.Baggage{},
				)
				result = bmock.NewResult()
				cmd    = bmock.NewCmd().RegisterExec(
					func(ctx context.Context, seq core.Seq, at time.Time, receiver any,
						proxy core.Proxy) (err error) {
						_, err = proxy.Send(0, result)
						return
					},
				)
				wantSpanName  = "Invoke " + internal_semconv.TypeStr(cmd)
				wantTimestamp = time.Now()
				addSpanAttrs  = []attribute.KeyValue{
					{Key: "cmd", Value: attribute.StringValue("cmd_value")},
				}
				addResultEventAttrs = []attribute.KeyValue{
					{Key: "result", Value: attribute.StringValue("result_value")},
				}
				addCmdMetricAttrs = []attribute.KeyValue{
					{Key: "cmdmetric", Value: attribute.StringValue("cmdmetric_value")},
				}
				addResultMetricAttrs = []attribute.KeyValue{
					{Key: "resultmetric", Value: attribute.StringValue("resultmetric_value")},
				}
				ops = []SetOption[any]{
					WithSpanStartOption[any](trace.WithTimestamp(wantTimestamp)),
					WithSpanAttributesFn(
						func(remoteAddr net.Addr, sentCmd hooks.SentCmd[any]) []attribute.KeyValue {
							asserterror.Equal(remoteAddr, net.Addr(wantAddr), t)
							asserterror.Equal(sentCmd.Seq, CmdSeq, t)
							asserterror.Equal(sentCmd.Size, CmdSize, t)
							asserterror.Equal(sentCmd.Cmd, core.Cmd[any](cmd), t)
							return addSpanAttrs
						},
					),
					WithSpanResultEventAttributesFn(
						func(sentCmd hooks.SentCmd[any], recvResult hooks.ReceivedResult) []attribute.KeyValue {
							asserterror.Equal(sentCmd.Seq, CmdSeq, t)
							asserterror.Equal(sentCmd.Size, CmdSize, t)
							asserterror.Equal(sentCmd.Cmd, core.Cmd[any](cmd), t)
							asserterror.Equal(recvResult.Seq, ResultSeq, t)
							asserterror.Equal(recvResult.Size, ResultSize, t)
							asserterror.Equal(recvResult.Result, core.Result(result), t)
							return addResultEventAttrs
						},
					),
					WithCmdMetricAttributesFn(
						func(sentCmd hooks.SentCmd[any], status semconv.CmdStreamCommandStatus, elapsedTime float64) []attribute.KeyValue {
							asserterror.Equal(sentCmd.Seq, CmdSeq, t)
							asserterror.Equal(sentCmd.Size, CmdSize, t)
							asserterror.Equal(sentCmd.Cmd, core.Cmd[any](cmd), t)
							asserterror.Equal(status, semconv.Ok, t)
							// asserterror.Equal(elapsedTime, wantElapsedTime, t)
							return addCmdMetricAttrs
						},
					),
					WithResultMetricAttributesFn(
						func(sentCmd hooks.SentCmd[any], recvResult hooks.ReceivedResult, elapsedTime float64) []attribute.KeyValue {
							asserterror.Equal(sentCmd.Seq, CmdSeq, t)
							asserterror.Equal(sentCmd.Size, CmdSize, t)
							asserterror.Equal(sentCmd.Cmd, core.Cmd[any](cmd), t)
							asserterror.Equal(recvResult.Seq, ResultSeq, t)
							asserterror.Equal(recvResult.Size, ResultSize, t)
							asserterror.Equal(recvResult.Result, core.Result(result), t)
							// asserterror.Equal(elapsedTime, wantElapsedTime, t)
							return addResultMetricAttrs
						},
					),
				}
			)
			otel.SetMeterProvider(meterProvider)
			otel.SetTracerProvider(tracerProvider)
			otel.SetTextMapPropagator(propagator)

			want := newWantVals(wantAddr, wantSpanName, cmd, result, semconv.Ok,
				[]trace.SpanStartOption{trace.WithTimestamp(wantTimestamp)},
				addSpanAttrs,
				addResultEventAttrs,
				addCmdMetricAttrs,
				addResultMetricAttrs,
				true,
			)
			testInvoke(want, meterProvider, tracerProvider, cmd, result, ops, t)
		})

}

func testInvoke(want wantVals,
	// wantAddr *net.TCPAddr, wantSpanName string,
	meterProvider mock.MeterProvider,
	tracerProvider mock.TracerProvider,
	cmd core.Cmd[any],
	result core.Result,
	ops []SetOption[any],
	t *testing.T,
) {
	// Invoke method do the following:
	// 0. Initialize metric vars.
	vars := mockServerMeterProvider(meterProvider, t)

	// 1. Define a regular cmd and result.
	// cmd and result are received as parameters.

	// 2. Start span.
	var (
		ctxWithSpan, span = mockTracerProviderForTraceCmd(tracerProvider, want.spanName,
			want.spanStartConfig, t)
	)

	// 3. Set span attributes.
	mockSpanAttributes(span, want.addr, want.spanAttrs, t)

	// 4. Invoke cmd with wrapped Proxy.
	var (
		invoker = hmock.NewInvoker[any]()
		proxy   = mockInvoker(invoker, want.addr, cmd, t)
	)

	// 5.1. Proxy should set span result-event attributes.
	mockSpanEvent(span, result, want.resultEventConfig, t)

	// 5.2. Proxy should record cmd and result metrics.
	mockMetricVars(ctxWithSpan, vars, want, t)

	// 6. End span.
	span.RegisterEnd(func(options ...trace.SpanEndOption) {})

	err := NewInvoker(invoker, ops...).Invoke(context.Background(), CmdSeq,
		time.Now(), CmdSize, cmd, proxy)
	asserterror.EqualError(err, nil, t)
}

func mockServerMeterProvider(meterProvider mock.MeterProvider, t *testing.T) (
	vars metricVars) {
	vars, fn1, fn2, fn3, fn4, fn5, fn6 := meterFns(t)
	meter := mock.NewMeter().RegisterInt64Counter(fn1).
		RegisterInt64Histogram(fn2).
		RegisterFloat64Histogram(fn3).
		RegisterInt64Counter(fn4).
		RegisterInt64Histogram(fn5).
		RegisterFloat64Histogram(fn6)
	meterProvider.RegisterMeter(
		func(name string, opts ...metric.MeterOption) metric.Meter {
			// TODO
			return meter
		},
	)
	return
}

func mockTracerProviderForTraceCmd(tracerProvider mock.TracerProvider, wantSpanName string,
	wantSpanStartConfig trace.SpanConfig, t *testing.T) (spanCtx context.Context, span mock.Span) {
	span = mock.NewSpan().RegisterSpanContext(
		func() trace.SpanContext {
			return spanContextFromTraceparent(Traceparent)
		},
	)
	spanCtx = mockTracerProvider(tracerProvider, wantSpanName, wantSpanStartConfig,
		span, t)
	return
}

func mockTracerProviderForRegularCmd(tracerProvider mock.TracerProvider, wantSpanName string,
	wantSpanStartConfig trace.SpanConfig, t *testing.T) (spanCtx context.Context, span mock.Span) {
	span = mock.NewSpan()
	spanCtx = mockTracerProvider(tracerProvider, wantSpanName, wantSpanStartConfig,
		span, t)
	return
}

func mockTracerProvider(tracerProvider mock.TracerProvider, wantSpanName string,
	wantSpanStartConfig trace.SpanConfig, span mock.Span, t *testing.T) (spanCtx context.Context) {
	spanCtx = trace.ContextWithSpan(context.Background(), span)
	tracer := mock.NewTracer().RegisterStart(
		func(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (
			context.Context, trace.Span) {
			// TODO check ctx, contains traceparent
			asserterror.Equal(spanName, wantSpanName, t)

			config := trace.NewSpanStartConfig(opts...)
			asserterror.EqualDeep(wantSpanStartConfig, config, t)

			return spanCtx, span
		},
	)

	tracerProvider.RegisterTracer(
		func(name string, options ...trace.TracerOption) trace.Tracer {
			asserterror.Equal(name, ScopeName, t)

			config := trace.NewTracerConfig(options...)
			asserterror.Equal(config.InstrumentationVersion(), Version(), t)

			return tracer
		},
	)
	return
}

func mockSpanAttributes(span mock.Span, wantAddr *net.TCPAddr,
	wantSpanAttrs []attribute.KeyValue, t *testing.T) {
	span.RegisterSetAttributes(
		func(attrs ...attribute.KeyValue) {
			asserterror.EqualDeep(attrs, wantSpanAttrs, t)
		},
	)
}

func mockInvoker(invoker hmock.Invoker[any], wantAddr *net.TCPAddr,
	wantCmd core.Cmd[any], t *testing.T) (proxy bmock.Proxy) {
	proxy = bmock.NewProxy().RegisterRemoteAddr(
		func() (addr net.Addr) { return wantAddr },
	)
	proxy.RegisterSend(
		func(seq core.Seq, result core.Result) (n int, err error) {
			return ResultSize, nil
		},
	)
	invoker.RegisterInvoke(
		func(ctx context.Context, seq core.Seq, at time.Time, bytesRead int,
			cmd core.Cmd[any], proxy core.Proxy) (err error) {
			asserterror.Equal(cmd, wantCmd, t)
			return cmd.Exec(ctx, seq, at, struct{}{}, proxy)
		},
	)
	return
}

func mockSpanEvent(span mock.Span, result core.Result, wantConfig trace.EventConfig,
	t *testing.T) {
	span.RegisterAddEvent(
		func(name string, options ...trace.EventOption) {
			asserterror.Equal(name, internal_semconv.ResultEventName, t)
			config := trace.NewEventConfig(options...)
			asserterror.EqualDeep(config.Attributes(), wantConfig.Attributes(), t)
		},
	)
}

func mockMetricVars(ctxWithSpan context.Context, vars metricVars, want wantVals,
	t *testing.T) {
	mockResultMetricVars(ctxWithSpan, vars, want, t)
	mockCmdMetricVars(ctxWithSpan, vars, want, t)
}

func mockResultMetricVars(ctxWithSpan context.Context, vars metricVars,
	want wantVals, t *testing.T) {
	rfn1, rfn2, rfn3 := resultMetricFns(ctxWithSpan, 1, ResultSize, want, t)
	vars.resultInt64Counter.RegisterAdd(rfn1)
	vars.resultInt64Histogram.RegisterRecord(rfn2)
	vars.resultFloat64Histogram.RegisterRecord(rfn3)
}

func mockCmdMetricVars(ctxWithSpan context.Context, vars metricVars,
	want wantVals, t *testing.T) {
	cfn1, cfn2, cfn3 := commandMetricFns(ctxWithSpan, 1, CmdSize, want, t)
	vars.cmdInt64Counter.RegisterAdd(cfn1)
	vars.cmdInt64Histogram.RegisterRecord(cfn2)
	vars.cmdFloat64Histogram.RegisterRecord(cfn3)
}

func resultMetricFns(wantCtx context.Context, wantIncr, wantResultSize int64,
	want wantVals,
	t *testing.T,
) (fn1 mock.AddFn, fn2 mock.RecordFn, fn3 mock.RecordFloatFn) {
	fn1 = func(ctx context.Context, incr int64, options ...metric.AddOption) {
		asserterror.Equal(ctx, wantCtx, t)
		asserterror.Equal(incr, wantIncr, t)
		config := metric.NewAddConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.resultMetricAddConfig.Attributes(), t)
	}
	fn2 = func(ctx context.Context, incr int64, options ...metric.RecordOption) {
		asserterror.Equal(ctx, wantCtx, t)
		asserterror.Equal(incr, wantResultSize, t)
		config := metric.NewRecordConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.resultMetricRecordConfig.Attributes(), t)
	}
	fn3 = func(ctx context.Context, incr float64, options ...metric.RecordOption) {
		asserterror.Equal(ctx, wantCtx, t)
		// asserterror.Equal(incr, float64(elapsedTime), t)
		config := metric.NewRecordConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.resultMetricRecordConfig.Attributes(), t)
	}
	return
}

func commandMetricFns(wantCtx context.Context, wantIncr, wantCmdSize int64,
	want wantVals,
	t *testing.T,
) (fn1 mock.AddFn, fn2 mock.RecordFn, fn3 mock.RecordFloatFn) {
	fn1 = func(ctx context.Context, incr int64, options ...metric.AddOption) {
		asserterror.Equal(ctx, wantCtx, t)
		asserterror.Equal(incr, wantIncr, t)
		config := metric.NewAddConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.cmdMetricAddConfig.Attributes(), t)
	}
	fn2 = func(ctx context.Context, incr int64, options ...metric.RecordOption) {
		asserterror.Equal(ctx, wantCtx, t)
		asserterror.Equal(incr, wantCmdSize, t)
		config := metric.NewRecordConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.cmdMetricRecordConfig.Attributes(), t)
	}
	fn3 = func(ctx context.Context, incr float64, options ...metric.RecordOption) {
		asserterror.Equal(ctx, wantCtx, t)
		// asserterror.Equal(incr, float64(elapsedTime), t)
		config := metric.NewRecordConfig(options)
		asserterror.EqualDeep(config.Attributes(), want.cmdMetricRecordConfig.Attributes(), t)
	}
	return
}

func meterFns(t *testing.T) (vars metricVars, fn1 mock.Int64CounterFn,
	fn2 mock.Int64HistogramFn,
	fn3 mock.Float64HistogramFn,
	fn4 mock.Int64CounterFn,
	fn5 mock.Int64HistogramFn,
	fn6 mock.Float64HistogramFn,
) {
	vars.cmdInt64Counter = mock.NewInt64Counter()
	fn1 = func(name string, options ...metric.Int64CounterOption) (c metric.Int64Counter, err error) {
		asserterror.Equal(name, semconv.CmdStreamServerCommandCountName, t)
		var (
			wantConf = metric.NewInt64CounterConfig(
				metric.WithUnit(semconv.CmdStreamServerCommandCountUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandCountDescription),
			)
			conf = metric.NewInt64CounterConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdInt64Counter, nil
	}

	vars.cmdInt64Histogram = mock.NewInt64Histogram()
	fn2 = func(name string, options ...metric.Int64HistogramOption) (h metric.Int64Histogram, err error) {
		asserterror.Equal(name, semconv.CmdStreamServerCommandSizeName, t)
		var (
			wantConf = metric.NewInt64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamServerCommandSizeUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandSizeDescription),
			)
			conf = metric.NewInt64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdInt64Histogram, nil
	}

	vars.cmdFloat64Histogram = mock.NewFloat64Histogram()
	fn3 = func(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
		asserterror.Equal(name, semconv.CmdStreamServerCommandDurationName, t)
		var (
			wantConf = metric.NewFloat64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamServerCommandDurationUnit),
				metric.WithDescription(semconv.CmdStreamServerCommandDurationDescription),
			)
			conf = metric.NewFloat64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdFloat64Histogram, nil
	}

	vars.resultInt64Counter = mock.NewInt64Counter()
	fn4 = func(name string, options ...metric.Int64CounterOption) (c metric.Int64Counter, err error) {
		asserterror.Equal(name, semconv.CmdStreamServerResultCountName, t)
		var (
			wantConf = metric.NewInt64CounterConfig(
				metric.WithUnit(semconv.CmdStreamServerResultCountUnit),
				metric.WithDescription(semconv.CmdStreamServerResultCountDescription),
			)
			conf = metric.NewInt64CounterConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultInt64Counter, nil
	}

	vars.resultInt64Histogram = mock.NewInt64Histogram()
	fn5 = func(name string, options ...metric.Int64HistogramOption) (h metric.Int64Histogram, err error) {
		asserterror.Equal(name, semconv.CmdStreamServerResultSizeName, t)
		var (
			wantConf = metric.NewInt64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamServerResultSizeUnit),
				metric.WithDescription(semconv.CmdStreamServerResultSizeDescription),
			)
			conf = metric.NewInt64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultInt64Histogram, nil
	}

	vars.resultFloat64Histogram = mock.NewFloat64Histogram()
	fn6 = func(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
		asserterror.Equal(name, semconv.CmdStreamServerResultDurationName, t)
		var (
			wantConf = metric.NewFloat64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamServerResultDurationUnit),
				metric.WithDescription(semconv.CmdStreamServerResultDurationDescription),
			)
			conf = metric.NewFloat64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultFloat64Histogram, nil
	}
	return
}

func newWantVals(addr *net.TCPAddr, spanName string, cmd core.Cmd[any],
	result core.Result,
	status semconv.CmdStreamCommandStatus,
	addSpanStartOptions []trace.SpanStartOption,
	addSpanAttrs []attribute.KeyValue,
	addResultEventAttrs []attribute.KeyValue,
	addCmdMetricAttrs []attribute.KeyValue,
	addResultMetricAttrs []attribute.KeyValue,
	server bool,
) wantVals {
	var (
		spanStartOptions = append([]trace.SpanStartOption{
			trace.WithSpanKind(trace.SpanKindServer)}, addSpanStartOptions...)
		spanAttrs = append(addSpanAttrs, []attribute.KeyValue{
			otel_semconv.NetworkPeerAddress(addr.IP.String()),
			otel_semconv.NetworkPeerPort(addr.Port),
			otel_semconv.NetworkProtocolName("tcp"),
		}...)
		resultEventOps = []trace.EventOption{
			trace.WithAttributes(addResultEventAttrs...),
		}
		cmdMetricAddConfig       = wantCmdMetricAddConfig(cmd, status, addCmdMetricAttrs)
		cmdMetricRecordConfig    = wantCmdMetricRecordConfig(cmd, status, addCmdMetricAttrs)
		resultMetricAddConfig    = wantResultMetricAddConfig(cmd, result, addResultMetricAttrs)
		resultMetricRecordConfig = wantResultMetricRecordConfig(cmd, result, addResultMetricAttrs)
	)
	if server {
		resultEventOps = append(resultEventOps,
			trace.WithAttributes(
				semconv.CmdStreamResultSeqKey.Int64(ResultSeq),
				semconv.CmdStreamResultSizeKey.Int64(ResultSize),
				semconv.CmdStreamResultTypeKey.String(internal_semconv.TypeStr(result)),
			))
	}

	return wantVals{
		addr:                     addr,
		spanName:                 spanName,
		spanStartConfig:          trace.NewSpanStartConfig(spanStartOptions...),
		spanAttrs:                spanAttrs,
		resultEventConfig:        trace.NewEventConfig(resultEventOps...),
		cmdMetricAddConfig:       cmdMetricAddConfig,
		cmdMetricRecordConfig:    cmdMetricRecordConfig,
		resultMetricAddConfig:    resultMetricAddConfig,
		resultMetricRecordConfig: resultMetricRecordConfig,
	}
}

type wantVals struct {
	addr                     *net.TCPAddr
	spanName                 string
	spanStartConfig          trace.SpanConfig
	spanAttrs                []attribute.KeyValue
	resultEventConfig        trace.EventConfig
	cmdMetricAddConfig       metric.AddConfig
	cmdMetricRecordConfig    metric.RecordConfig
	resultMetricAddConfig    metric.AddConfig
	resultMetricRecordConfig metric.RecordConfig
}

func wantResultMetricAddConfig(cmd core.Cmd[any],
	result core.Result,
	addAttrs []attribute.KeyValue,
) metric.AddConfig {
	return metric.NewAddConfig(
		[]metric.AddOption{
			metric.WithAttributeSet(
				attribute.NewSet(
					append(
						addAttrs,
						semconv.CmdStreamCommandTypeKey.String(internal_semconv.TypeStr(cmd)),
						semconv.CmdStreamResultTypeKey.String(internal_semconv.TypeStr(result)),
					)...,
				)),
		},
	)
}

func wantResultMetricRecordConfig(cmd core.Cmd[any],
	result core.Result,
	addAttrs []attribute.KeyValue,
) metric.RecordConfig {
	return metric.NewRecordConfig(
		[]metric.RecordOption{
			metric.WithAttributeSet(
				attribute.NewSet(
					append(
						addAttrs,
						semconv.CmdStreamCommandTypeKey.String(internal_semconv.TypeStr(cmd)),
						semconv.CmdStreamResultTypeKey.String(internal_semconv.TypeStr(result)),
					)...,
				)),
		},
	)
}

func wantCmdMetricAddConfig(cmd core.Cmd[any],
	status semconv.CmdStreamCommandStatus,
	addAttrs []attribute.KeyValue,
) metric.AddConfig {
	return metric.NewAddConfig(
		[]metric.AddOption{
			metric.WithAttributeSet(
				attribute.NewSet(
					append(
						addAttrs,
						semconv.CmdStreamCommandTypeKey.String(internal_semconv.TypeStr(cmd)),
						semconv.CmdStreamCommandStatusKey.String(string(status)),
					)...,
				)),
		},
	)
}

func wantCmdMetricRecordConfig(cmd core.Cmd[any],
	status semconv.CmdStreamCommandStatus,
	addAttrs []attribute.KeyValue,
) metric.RecordConfig {
	return metric.NewRecordConfig(
		[]metric.RecordOption{
			metric.WithAttributeSet(
				attribute.NewSet(
					append(
						addAttrs,
						semconv.CmdStreamCommandTypeKey.String(internal_semconv.TypeStr(cmd)),
						semconv.CmdStreamCommandStatusKey.String(string(status)),
					)...,
				)),
		},
	)
}

func spanContextFromTraceparent(traceparent string) trace.SpanContext {
	carrier := propagation.MapCarrier{
		"traceparent": traceparent,
	}

	ctx := propagation.TraceContext{}.Extract(context.Background(), carrier)
	return trace.SpanContextFromContext(ctx)
}
