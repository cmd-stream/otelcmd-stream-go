package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/metric"
)

type RecordFloatFn func(ctx context.Context, incr float64, options ...metric.RecordOption)

func NewFloat64Histogram() Float64Histogram {
	return Float64Histogram{Mock: mok.New("Float64Histogram")}
}

type Float64Histogram struct {
	*mok.Mock
	metric.Float64Histogram
}

func (c Float64Histogram) RegisterRecord(fn RecordFloatFn) Float64Histogram {
	c.Register("Record", fn)
	return c
}

func (c Float64Histogram) Record(ctx context.Context, incr float64,
	options ...metric.RecordOption) {
	_, err := c.Call("Record", ctx, incr, options)
	if err != nil {
		panic(err)
	}
}
