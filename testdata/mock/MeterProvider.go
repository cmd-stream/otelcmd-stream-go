package mock

import (
	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/metric"
)

type MeterFn func(name string, opts ...metric.MeterOption) metric.Meter

func NewMeterProvider() MeterProvider {
	return MeterProvider{Mock: mok.New("MeterProvider")}
}

type MeterProvider struct {
	*mok.Mock
	metric.MeterProvider
}

func (p MeterProvider) RegisterMeter(fn MeterFn) MeterProvider {
	p.Register("Meter", fn)
	return p
}

func (p MeterProvider) Meter(name string, opts ...metric.MeterOption) metric.Meter {
	results, err := p.Call("Meter", name, opts)
	if err != nil {
		panic(err)
	}
	m, _ := results[0].(metric.Meter)
	return m
}
