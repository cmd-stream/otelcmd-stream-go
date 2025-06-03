package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/trace"
)

type StartFn func(ctx context.Context, spanName string,
	opts ...trace.SpanStartOption) (context.Context, trace.Span)

func NewTracer() Tracer {
	return Tracer{Mock: mok.New("Tracer")}
}

type Tracer struct {
	trace.Tracer
	*mok.Mock
}

func (t Tracer) Start(ctx context.Context, spanName string,
	opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	results, err := t.Call("Start", ctx, spanName, opts)
	if err != nil {
		panic(err)
	}
	actx, _ := results[0].(context.Context)
	span, _ := results[1].(trace.Span)
	return actx, span
}

func (t Tracer) RegisterStart(fn StartFn) Tracer {
	t.Register("Start", fn)
	return t
}
