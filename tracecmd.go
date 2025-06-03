package otelcs

import (
	"context"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/otelcmd-stream-go/internal/semconv"
)

type traceCmd[T any] interface {
	SetCarrier(carrier map[string]string)
	Carrier() map[string]string
	InnerCmd() core.Cmd[T]
}

// NewTraceCmd creates a new TraceCmd.
func NewTraceCmd[T any, V core.Cmd[T]](cmd V) TraceCmd[T, V] {
	return TraceCmd[T, V]{
		MapCarrier: new(map[string]string),
		Cmd:        cmd,
	}
}

// TraceCmd wraps a core.Cmd and carries tracing context for propagation.
type TraceCmd[T any, V core.Cmd[T]] struct {
	MapCarrier *map[string]string
	Cmd        V
}

func (c TraceCmd[T, V]) Exec(ctx context.Context, seq core.Seq, at time.Time,
	receiver T, proxy core.Proxy) error {
	return c.Cmd.Exec(ctx, seq, at, receiver, proxy)
}

func (c TraceCmd[T, V]) TypeStr() string {
	return semconv.TypeStr(c.Cmd) + " (trace)"
}

func (c TraceCmd[T, V]) SetCarrier(carrier map[string]string) {
	if c.MapCarrier == nil {
		panic("TraceCmd was not initialized with NewTraceCmd")
	}
	*c.MapCarrier = carrier
}

func (c TraceCmd[T, V]) Carrier() (carrier map[string]string) {
	if c.MapCarrier == nil {
		return
	}
	return *c.MapCarrier
}

func (c TraceCmd[T, V]) InnerCmd() core.Cmd[T] {
	return c.Cmd
}
