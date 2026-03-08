package mock

import (
	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/metric"
)

type Int64CounterFn func(name string, options ...metric.Int64CounterOption) (metric.Int64Counter, error)
type Int64HistogramFn func(name string, options ...metric.Int64HistogramOption) (metric.Int64Histogram, error)
type Float64HistogramFn func(name string, options ...metric.Float64HistogramOption) (metric.Float64Histogram, error)

func NewMeter() Meter {
	return Meter{Mock: mok.New("Meter")}
}

type Meter struct {
	*mok.Mock
	metric.Meter
}

func (m Meter) RegisterInt64Counter(fn Int64CounterFn) Meter {
	m.Register("Int64Counter", fn)
	return m
}

func (m Meter) RegisterInt64Histogram(fn Int64HistogramFn) Meter {
	m.Register("Int64Histogram", fn)
	return m
}

func (m Meter) RegisterFloat64Histogram(fn Float64HistogramFn) Meter {
	m.Register("Float64Histogram", fn)
	return m
}

func (m Meter) Int64Counter(name string, options ...metric.Int64CounterOption) (c metric.Int64Counter, err error) {
	results, err := m.Call("Int64Counter", name, options)
	if err != nil {
		panic(err)
	}
	c, _ = results[0].(metric.Int64Counter)
	err, _ = results[1].(error)
	return
}

func (m Meter) Int64Histogram(name string, options ...metric.Int64HistogramOption) (h metric.Int64Histogram, err error) {
	results, err := m.Call("Int64Histogram", name, options)
	if err != nil {
		panic(err)
	}
	h, _ = results[0].(metric.Int64Histogram)
	err, _ = results[1].(error)
	return
}

func (m Meter) Float64Histogram(name string, options ...metric.Float64HistogramOption) (h metric.Float64Histogram, err error) {
	results, err := m.Call("Float64Histogram", name, options)
	if err != nil {
		panic(err)
	}
	h, _ = results[0].(metric.Float64Histogram)
	err, _ = results[1].(error)
	return
}

func (m Meter) Int64UpDownCounter(name string, options ...metric.Int64UpDownCounterOption) (metric.Int64UpDownCounter, error) {
	panic("not implemented")
}

func (m Meter) Int64Gauge(name string, options ...metric.Int64GaugeOption) (metric.Int64Gauge, error) {
	panic("not implemented")
}

func (m Meter) Int64ObservableCounter(name string, options ...metric.Int64ObservableCounterOption) (metric.Int64ObservableCounter, error) {
	panic("not implemented")
}

func (m Meter) Int64ObservableUpDownCounter(name string, options ...metric.Int64ObservableUpDownCounterOption) (metric.Int64ObservableUpDownCounter, error) {
	panic("not implemented")
}

func (m Meter) Int64ObservableGauge(name string, options ...metric.Int64ObservableGaugeOption) (metric.Int64ObservableGauge, error) {
	panic("not implemented")
}

func (m Meter) Float64Counter(name string, options ...metric.Float64CounterOption) (metric.Float64Counter, error) {
	panic("not implemented")
}

func (m Meter) Float64UpDownCounter(name string, options ...metric.Float64UpDownCounterOption) (metric.Float64UpDownCounter, error) {
	panic("not implemented")
}

func (m Meter) Float64Gauge(name string, options ...metric.Float64GaugeOption) (metric.Float64Gauge, error) {
	panic("not implemented")
}

func (m Meter) Float64ObservableCounter(name string, options ...metric.Float64ObservableCounterOption) (metric.Float64ObservableCounter, error) {
	panic("not implemented")
}

func (m Meter) Float64ObservableUpDownCounter(name string, options ...metric.Float64ObservableUpDownCounterOption) (metric.Float64ObservableUpDownCounter, error) {
	panic("not implemented")
}

func (m Meter) Float64ObservableGauge(name string, options ...metric.Float64ObservableGaugeOption) (metric.Float64ObservableGauge, error) {
	panic("not implemented")
}

func (m Meter) RegisterCallback(f metric.Callback, instruments ...metric.Observable) (metric.Registration, error) {
	panic("not implemented")
}
