package mock

import (
	"context"

	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/metric"
)

type AddFn func(ctx context.Context, incr int64, options ...metric.AddOption)

func NewInt64Counter() Int64Counter {
	return Int64Counter{Mock: mok.New("Int64Counter")}
}

type Int64Counter struct {
	*mok.Mock
	metric.Int64Counter
}

func (c Int64Counter) RegisterAdd(fn AddFn) Int64Counter {
	c.Register("Add", fn)
	return c
}

func (c Int64Counter) Add(ctx context.Context, incr int64, options ...metric.AddOption) {
	_, err := c.Call("Add", ctx, incr, options)
	if err != nil {
		panic(err)
	}
}
