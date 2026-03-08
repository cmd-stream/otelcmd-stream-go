package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/metric"
)

type RecordFn func(ctx context.Context, incr int64, options ...metric.RecordOption)

func NewInt64Histogram() Int64Histogram {
	return Int64Histogram{Mock: mok.New("Int64Histogram")}
}

type Int64Histogram struct {
	*mok.Mock
	metric.Int64Histogram
}

func (c Int64Histogram) RegisterRecord(fn RecordFn) Int64Histogram {
	c.Register("Record", fn)
	return c
}

func (c Int64Histogram) Record(ctx context.Context, incr int64,
	options ...metric.RecordOption) {
	_, err := c.Call("Record", ctx, incr, options)
	if err != nil {
		panic(err)
	}
}
