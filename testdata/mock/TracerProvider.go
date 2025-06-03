package mock

import (
	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/trace"
)

type TracerFn func(name string, options ...trace.TracerOption) trace.Tracer

func NewTracerProvider() TracerProvider {
	return TracerProvider{Mock: mok.New("TracerProvider")}
}

type TracerProvider struct {
	trace.TracerProvider
	*mok.Mock
}

func (p TracerProvider) RegisterTracer(fn TracerFn) TracerProvider {
	p.Register("Tracer", fn)
	return p
}

func (p TracerProvider) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	results, err := p.Call("Tracer", name, options)
	if err != nil {
		panic(err)
	}
	t, _ := results[0].(trace.Tracer)
	return t
}
