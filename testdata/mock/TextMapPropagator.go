package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/propagation"
)

type InjectFn func(ctx context.Context, carrier propagation.TextMapCarrier)
type ExtractFn func(ctx context.Context,
	carrier propagation.TextMapCarrier) context.Context
type FieldsFn func() []string

func NewTextMapPropagator() TextMapPropagator {
	return TextMapPropagator{mok.New("TextMapPropagator")}
}

type TextMapPropagator struct {
	*mok.Mock
}

func (p TextMapPropagator) RegisterInject(fn InjectFn) TextMapPropagator {
	p.Register("Inject", fn)
	return p
}

func (p TextMapPropagator) RegisterExtract(fn ExtractFn) TextMapPropagator {
	p.Register("Extract", fn)
	return p
}

func (p TextMapPropagator) RegisterFields(fn FieldsFn) TextMapPropagator {
	p.Register("Fields", fn)
	return p
}

func (p TextMapPropagator) Inject(ctx context.Context,
	carrier propagation.TextMapCarrier) {
	_, err := p.Call("Inject", ctx, carrier)
	if err != nil {
		panic(err)
	}
}

func (p TextMapPropagator) Extract(ctx context.Context,
	carrier propagation.TextMapCarrier) context.Context {
	result, err := p.Call("Extract", ctx, carrier)
	if err != nil {
		panic(err)
	}
	actx, _ := result[0].(context.Context)
	return actx
}

func (p TextMapPropagator) Fields() []string {
	result, err := p.Call("Fields")
	if err != nil {
		panic(err)
	}
	return result[0].([]string)
}
