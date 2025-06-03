package otelcs

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/cmd-stream/core-go"
	bmock "github.com/cmd-stream/core-go/testdata/mock"
	internal_semconv "github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
	semconv "github.com/cmd-stream/otelcmd-stream-go/semconv"
	"github.com/cmd-stream/otelcmd-stream-go/testdata/mock"
	"github.com/cmd-stream/sender-go/hooks"
	asserterror "github.com/ymz-ncnk/assert/error"
	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/metric/noop"
	"go.opentelemetry.io/otel/propagation"
	otel_semconv "go.opentelemetry.io/otel/semconv/v1.22.0"
	"go.opentelemetry.io/otel/trace"
	tracenop "go.opentelemetry.io/otel/trace/noop"
)

func TestSendHooks(t *testing.T) {

	t.Run("BeforeSend", func(t *testing.T) {

		t.Run("Should work with TraceCmd", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantErr  error = nil
				cmd            = bmock.NewCmd()
				traceCmd       = NewTraceCmd(cmd)

				spanStartOptions = []trace.SpanStartOption{
					trace.WithSpanKind(trace.SpanKindClient)}
				wantSpanStartConfig = trace.NewSpanStartConfig(spanStartOptions...)

				tracerProvider = mock.NewTracerProvider()
				propagator     = propagation.NewCompositeTextMapPropagator(
					propagation.TraceContext{},
					propagation.Baggage{},
				)
			)

			otel.SetTracerProvider(tracerProvider)
			otel.SetTextMapPropagator(propagator)

			var (
				_, span = mockTracerProviderForTraceCmd(tracerProvider,
					defaultClientSpanNameFormatter(traceCmd), wantSpanStartConfig, t)
				mocks = []*mok.Mock{tracerProvider.Mock, span.Mock, cmd.Mock}
			)

			hooks := NewHooksFactory[any]().New()
			_, err := hooks.BeforeSend(context.Background(), traceCmd)
			asserterror.EqualError(err, wantErr, t)

			asserterror.EqualDeep(hooks.(*Hooks[any]).span, span, t)
			asserterror.EqualDeep(traceCmd.Carrier(),
				map[string]string{"traceparent": Traceparent}, t)
			// TODO hooks.(*SendHooks[any]).sendTime

			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

		t.Run("Should work with a regular Cmd", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantErr          error = nil
				cmd                    = bmock.NewCmd()
				spanStartOptions       = []trace.SpanStartOption{
					trace.WithSpanKind(trace.SpanKindClient)}
				wantSpanStartConfig = trace.NewSpanStartConfig(spanStartOptions...)

				tracerProvider = mock.NewTracerProvider()
			)
			otel.SetTracerProvider(tracerProvider)

			var (
				_, span = mockTracerProviderForRegularCmd(tracerProvider,
					defaultClientSpanNameFormatter(cmd), wantSpanStartConfig, t)
				mocks = []*mok.Mock{tracerProvider.Mock, span.Mock, cmd.Mock}
			)

			hooks := NewHooksFactory[any]().New()
			_, err := hooks.BeforeSend(context.Background(), cmd)
			asserterror.EqualError(err, wantErr, t)

			asserterror.EqualDeep(hooks.(*Hooks[any]).span, span, t)

			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

		t.Run("We should be able to set TracerProvider and Propagator with options",
			func(t *testing.T) {
				otel.SetTracerProvider(tracenop.NewTracerProvider())
				otel.SetMeterProvider(noop.NewMeterProvider())
				otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

				var (
					wantErr          error = nil
					cmd                    = bmock.NewCmd()
					traceCmd               = NewTraceCmd(cmd)
					spanStartOptions       = []trace.SpanStartOption{
						trace.WithSpanKind(trace.SpanKindClient)}
					wantSpanStartConfig = trace.NewSpanStartConfig(spanStartOptions...)

					tracerProvider = mock.NewTracerProvider()
					propagator     = propagation.NewCompositeTextMapPropagator(
						propagation.TraceContext{},
						propagation.Baggage{},
					)
				)

				ops := []SetOption[any]{
					WithTracerProvider[any](tracerProvider),
					WithPropagator[any](propagator),
				}

				var (
					_, span = mockTracerProviderForTraceCmd(tracerProvider,
						defaultClientSpanNameFormatter(traceCmd), wantSpanStartConfig, t)
					mocks = []*mok.Mock{tracerProvider.Mock, span.Mock, cmd.Mock}
				)

				hooks := NewHooksFactory[any](ops...).New()
				_, err := hooks.BeforeSend(context.Background(), traceCmd)
				asserterror.EqualError(err, wantErr, t)

				asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
			})

		t.Run("Should get a TracerProvider from the context", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantErr error = nil
				cmd           = bmock.NewCmd()
				// 00-8e4e4fc7a0b349d2b2039ab9c1e2a5f7-5d3f9a81cd8731be-01
				sc = trace.NewSpanContext(trace.SpanContextConfig{
					TraceID:    [16]byte([]byte("8e4e4fc7a0b349d2b2039ab9c1e2a5f7")),
					SpanID:     [8]byte([]byte("5d3f9a81cd8731be")),
					TraceFlags: trace.FlagsSampled,
					Remote:     false,
				})
				ctx = trace.ContextWithSpanContext(context.Background(), sc)

				mocks = []*mok.Mock{cmd.Mock}
			)

			hooks := NewHooksFactory[any]().New()
			_, err := hooks.BeforeSend(ctx, cmd)
			asserterror.EqualError(err, wantErr, t)

			asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
		})

	})

	t.Run("OnError", func(t *testing.T) {

		t.Run("Should work", func(t *testing.T) {
			testOnError(nil, t)
		})

		t.Run("We chould be able to add custom span attributes", func(t *testing.T) {
			addAttrs := []attribute.KeyValue{
				{Key: "cmd", Value: attribute.StringValue("cmd_value")},
			}
			testOnError(addAttrs, t)
		})

	})

	t.Run("OnResult", func(t *testing.T) {

		t.Run("Should work with last one result", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				meterProvider = mock.NewMeterProvider()

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				result = bmock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				recvResult = hooks.ReceivedResult{
					Seq:    ResultSeq,
					Size:   ResultSize,
					Result: result,
				}
				wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
				want         = newWantVals(wantAddr, wantSpanName, cmd, result,
					semconv.Ok, nil, nil, nil, nil, nil, false)

				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
				}

				span = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterEnd(
					func(options ...trace.SpanEndOption) {},
				)
				ctxWithSpan = trace.ContextWithSpan(context.Background(), span)
				vars        = mockClientMeterProvider(meterProvider, t)
			)
			mockMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnResult(ctxWithSpan, sentCmd, recvResult, nil, span, meterProvider,
				ops, t)
		})

		t.Run("Should work with not last one result", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				result = bmock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return false },
				)
				recvResult = hooks.ReceivedResult{
					Seq:    ResultSeq,
					Size:   ResultSize,
					Result: result,
				}
				wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
				want         = newWantVals(wantAddr, wantSpanName, cmd, result,
					semconv.Ok, nil, nil, nil, nil, nil, false)

				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
				}

				span = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				)
				ctxWithSpan   = trace.ContextWithSpan(context.Background(), span)
				meterProvider = mock.NewMeterProvider()
				vars          = mockClientMeterProvider(meterProvider, t)
			)
			mockResultMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnResult(ctxWithSpan, sentCmd, recvResult, nil, span, meterProvider,
				ops, t)
		})

		t.Run("Should work in case of an error", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				err          = errors.New("test error")
				wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
				want         = newWantVals(wantAddr, wantSpanName, cmd, bmock.NewResult(),
					semconv.Failed, nil, nil, nil, nil, nil, false)

				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
				}

				meterProvider = mock.NewMeterProvider()
				vars          = mockClientMeterProvider(meterProvider, t)
				span          = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := []attribute.KeyValue{otel_semconv.ErrorTypeKey.String(internal_semconv.TypeStr(err))}
						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetStatus(
					func(code codes.Code, description string) {
						asserterror.Equal(code, codes.Error, t)
						asserterror.Equal(description, err.Error(), t)
					},
				).RegisterEnd(
					func(options ...trace.SpanEndOption) {},
				)
				ctxWithSpan = trace.ContextWithSpan(context.Background(), span)
			)
			mockCmdMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnResult(ctxWithSpan, sentCmd, hooks.ReceivedResult{}, err, span,
				meterProvider, ops, t)
		})

		t.Run("We should be able to add span/metric attributes", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				result = bmock.NewResult().RegisterLastOne(
					func() (lastOne bool) { return true },
				)
				recvResult = hooks.ReceivedResult{
					Seq:    ResultSeq,
					Size:   ResultSize,
					Result: result,
				}
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
				want = newWantVals(wantAddr, wantSpanName, cmd, bmock.NewResult(),
					semconv.Ok,
					[]trace.SpanStartOption{trace.WithTimestamp(wantTimestamp)},
					addSpanAttrs,
					addResultEventAttrs,
					addCmdMetricAttrs,
					addResultMetricAttrs,
					false,
				)

				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
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

				meterProvider = mock.NewMeterProvider()
				vars          = mockClientMeterProvider(meterProvider, t)
				span          = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterEnd(
					func(options ...trace.SpanEndOption) {},
				)
				ctxWithSpan = trace.ContextWithSpan(context.Background(), span)
			)
			mockSpanEvent(span, result, want.resultEventConfig, t)
			mockMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnResult(ctxWithSpan, sentCmd, recvResult, nil, span, meterProvider,
				ops, t)
		})

	})

	t.Run("OnTimeout", func(t *testing.T) {

		t.Run("Should work", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				err          = errors.New("test error")
				wantSpanName = "Invoke " + internal_semconv.TypeStr(cmd)
				want         = newWantVals(wantAddr, wantSpanName, cmd, bmock.NewResult(),
					semconv.Timeout, nil, nil, nil, nil, nil, false)

				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
				}

				meterProvider = mock.NewMeterProvider()
				vars          = mockClientMeterProvider(meterProvider, t)
				span          = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := []attribute.KeyValue{otel_semconv.ErrorTypeKey.String(internal_semconv.TypeStr(err))}
						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetStatus(
					func(code codes.Code, description string) {
						asserterror.Equal(code, codes.Error, t)
						asserterror.Equal(description, err.Error(), t)
					},
				).RegisterEnd(
					func(options ...trace.SpanEndOption) {},
				)
				ctxWithSpan = trace.ContextWithSpan(context.Background(), span)
			)
			mockCmdMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnTimeout(ctxWithSpan, sentCmd, err, span, meterProvider, ops, t)
		})

		t.Run("We should be able to add own span/metric attributes", func(t *testing.T) {
			otel.SetTracerProvider(tracenop.NewTracerProvider())
			otel.SetMeterProvider(noop.NewMeterProvider())
			otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator())

			var (
				wantAddr = &net.TCPAddr{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 8080,
				}

				cmd     = bmock.NewCmd()
				sentCmd = hooks.SentCmd[any]{
					Seq:  CmdSeq,
					Size: CmdSize,
					Cmd:  cmd,
				}
				err           = errors.New("test error")
				wantSpanName  = "Invoke " + internal_semconv.TypeStr(cmd)
				wantTimestamp = time.Now()
				addSpanAttrs  = []attribute.KeyValue{
					{Key: "cmd", Value: attribute.StringValue("cmd_value")},
				}
				addCmdMetricAttrs = []attribute.KeyValue{
					{Key: "cmdmetric", Value: attribute.StringValue("cmdmetric_value")},
				}
				want = newWantVals(wantAddr, wantSpanName, cmd, bmock.NewResult(),
					semconv.Timeout,
					[]trace.SpanStartOption{trace.WithTimestamp(wantTimestamp)},
					addSpanAttrs,
					nil,
					addCmdMetricAttrs,
					nil,
					false)
				ops = []SetOption[any]{
					WithServerAddr[any](wantAddr),
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
					WithCmdMetricAttributesFn(
						func(sentCmd hooks.SentCmd[any], status semconv.CmdStreamCommandStatus, elapsedTime float64) []attribute.KeyValue {
							asserterror.Equal(sentCmd.Seq, CmdSeq, t)
							asserterror.Equal(sentCmd.Size, CmdSize, t)
							asserterror.Equal(sentCmd.Cmd, core.Cmd[any](cmd), t)
							asserterror.Equal(status, semconv.Timeout, t)
							// asserterror.Equal(elapsedTime, wantElapsedTime, t)
							return addCmdMetricAttrs
						},
					),
				}

				meterProvider = mock.NewMeterProvider()
				vars          = mockClientMeterProvider(meterProvider, t)
				span          = mock.NewSpan().RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := []attribute.KeyValue{otel_semconv.ErrorTypeKey.String(internal_semconv.TypeStr(err))}
						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetAttributes(
					func(attrs ...attribute.KeyValue) {
						wantAttrs := make([]attribute.KeyValue, 0, len(want.spanAttrs)+2)
						wantAttrs = append(wantAttrs, want.spanAttrs...)
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
						wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

						asserterror.EqualDeep(attrs, wantAttrs, t)
					},
				).RegisterSetStatus(
					func(code codes.Code, description string) {
						asserterror.Equal(code, codes.Error, t)
						asserterror.Equal(description, err.Error(), t)
					},
				).RegisterEnd(
					func(options ...trace.SpanEndOption) {},
				)
				ctxWithSpan = trace.ContextWithSpan(context.Background(), span)
			)
			mockCmdMetricVars(ctxWithSpan, vars, want, t)
			otel.SetMeterProvider(meterProvider)

			testOnTimeout(ctxWithSpan, sentCmd, err, span, meterProvider, ops, t)
		})

	})

}

func testOnError(addAttrs []attribute.KeyValue, t *testing.T) {
	var (
		wantAddr = &net.TCPAddr{
			IP:   net.ParseIP("127.0.0.1"),
			Port: 8080,
		}
		err     error = errors.New("some error")
		sentCmd       = hooks.SentCmd[any]{
			Seq:  10,
			Size: 100,
			Cmd:  bmock.NewCmd(),
		}
		span = mock.NewSpan().RegisterSetAttributes(
			func(attrs ...attribute.KeyValue) {
				kv := otel_semconv.ErrorTypeKey.String(internal_semconv.TypeStr(err))
				asserterror.EqualDeep([]attribute.KeyValue{kv}, attrs, t)
			},
		).RegisterSetAttributes(
			func(attrs ...attribute.KeyValue) {
				addrAttrs := internal_semconv.AddrAttrs(wantAddr)
				wantAttrs := make([]attribute.KeyValue, 0, len(addAttrs)+len(addrAttrs)+2)

				wantAttrs = append(wantAttrs, addAttrs...)
				wantAttrs = append(wantAttrs, addrAttrs...)
				wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSeqKey.Int64(int64(sentCmd.Seq)))
				wantAttrs = append(wantAttrs, semconv.CmdStreamCommandSizeKey.Int64(int64(sentCmd.Size)))

				asserterror.EqualDeep(attrs, wantAttrs, t)
			},
		).RegisterSetStatus(
			func(code codes.Code, description string) {
				asserterror.Equal(code, codes.Error, t)
				asserterror.Equal(description, err.Error(), t)
			},
		).RegisterEnd(
			func(options ...trace.SpanEndOption) {},
		)
	)
	hooks := NewHooksFactory[any](
		WithServerAddr[any](wantAddr),
		WithSpanAttributesFn(
			func(remoteAddr net.Addr, sentCmd hooks.SentCmd[any]) []attribute.KeyValue {
				return addAttrs
			},
		),
	).New()
	hooks.(*Hooks[any]).span = span

	hooks.OnError(context.Background(), sentCmd, err)
}

func testOnResult(ctxWithSpan context.Context, sentCmd hooks.SentCmd[any],
	recvResult hooks.ReceivedResult,
	err error,
	span mock.Span,
	meterProvider mock.MeterProvider,
	ops []SetOption[any],
	t *testing.T,
) {
	mocks := []*mok.Mock{sentCmd.Cmd.(bmock.Cmd).Mock, meterProvider.Mock, span.Mock}
	if recvResult.Result != nil {
		mocks = append(mocks, recvResult.Result.(bmock.Result).Mock)
	}
	hooks := NewHooksFactory[any](ops...).New()
	hooks.(*Hooks[any]).span = span

	hooks.OnResult(ctxWithSpan, sentCmd, recvResult, err)

	asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
}

func testOnTimeout(ctxWithSpan context.Context, sentCmd hooks.SentCmd[any],
	err error,
	span mock.Span,
	meterProvider mock.MeterProvider,
	ops []SetOption[any],
	t *testing.T) {
	mocks := []*mok.Mock{sentCmd.Cmd.(bmock.Cmd).Mock, meterProvider.Mock, span.Mock}
	hooks := NewHooksFactory[any](ops...).New()
	hooks.(*Hooks[any]).span = span

	hooks.OnTimeout(ctxWithSpan, sentCmd, err)

	asserterror.EqualDeep(mok.CheckCalls(mocks), mok.EmptyInfomap, t)
}

func mockClientMeterProvider(meterProvider mock.MeterProvider, t *testing.T) (
	vars metricVars) {
	vars, fn1, fn2, fn3, fn4, fn5, fn6 := clientMeterFns(t)
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

func clientMeterFns(t *testing.T) (vars metricVars, fn1 mock.Int64CounterFn,
	fn2 mock.Int64HistogramFn,
	fn3 mock.Float64HistogramFn,
	fn4 mock.Int64CounterFn,
	fn5 mock.Int64HistogramFn,
	fn6 mock.Float64HistogramFn,
) {
	vars.cmdInt64Counter = mock.NewInt64Counter()
	fn1 = func(name string, options ...metric.Int64CounterOption) (c metric.Int64Counter, err error) {
		asserterror.Equal(name, semconv.CmdStreamClientCommandCountName, t)
		var (
			wantConf = metric.NewInt64CounterConfig(
				metric.WithUnit(semconv.CmdStreamClientCommandCountUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandCountDescription),
			)
			conf = metric.NewInt64CounterConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdInt64Counter, nil
	}

	vars.cmdInt64Histogram = mock.NewInt64Histogram()
	fn2 = func(name string, options ...metric.Int64HistogramOption) (h metric.Int64Histogram, err error) {
		asserterror.Equal(name, semconv.CmdStreamClientCommandSizeName, t)
		var (
			wantConf = metric.NewInt64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamClientCommandSizeUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandSizeDescription),
			)
			conf = metric.NewInt64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdInt64Histogram, nil
	}

	vars.cmdFloat64Histogram = mock.NewFloat64Histogram()
	fn3 = func(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
		asserterror.Equal(name, semconv.CmdStreamClientCommandDurationName, t)
		var (
			wantConf = metric.NewFloat64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamClientCommandDurationUnit),
				metric.WithDescription(semconv.CmdStreamClientCommandDurationDescription),
			)
			conf = metric.NewFloat64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.cmdFloat64Histogram, nil
	}

	vars.resultInt64Counter = mock.NewInt64Counter()
	fn4 = func(name string, options ...metric.Int64CounterOption) (c metric.Int64Counter, err error) {
		asserterror.Equal(name, semconv.CmdStreamClientResultCountName, t)
		var (
			wantConf = metric.NewInt64CounterConfig(
				metric.WithUnit(semconv.CmdStreamClientResultCountUnit),
				metric.WithDescription(semconv.CmdStreamClientResultCountDescription),
			)
			conf = metric.NewInt64CounterConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultInt64Counter, nil
	}

	vars.resultInt64Histogram = mock.NewInt64Histogram()
	fn5 = func(name string, options ...metric.Int64HistogramOption) (h metric.Int64Histogram, err error) {
		asserterror.Equal(name, semconv.CmdStreamClientResultSizeName, t)
		var (
			wantConf = metric.NewInt64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamClientResultSizeUnit),
				metric.WithDescription(semconv.CmdStreamClientResultSizeDescription),
			)
			conf = metric.NewInt64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultInt64Histogram, nil
	}

	vars.resultFloat64Histogram = mock.NewFloat64Histogram()
	fn6 = func(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error) {
		asserterror.Equal(name, semconv.CmdStreamClientResultDurationName, t)
		var (
			wantConf = metric.NewFloat64HistogramConfig(
				metric.WithUnit(semconv.CmdStreamClientResultDurationUnit),
				metric.WithDescription(semconv.CmdStreamClientResultDurationDescription),
			)
			conf = metric.NewFloat64HistogramConfig(options...)
		)
		asserterror.EqualDeep(conf, wantConf, t)
		return vars.resultFloat64Histogram, nil
	}
	return
}
