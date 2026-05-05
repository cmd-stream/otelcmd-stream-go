package otelcmd

import (
	"net"
	"time"

	"github.com/cmd-stream/cmd-stream-go/core"
	"github.com/cmd-stream/cmd-stream-go/sender/hooks"
)

type ProxyCallbackFn func(recvResult hooks.ReceivedResult)

func NewProxy[T any](proxy core.Proxy, callback ProxyCallbackFn) *Proxy[T] {
	return &Proxy[T]{proxy: proxy, callback: callback}
}

type Proxy[T any] struct {
	proxy     core.Proxy
	callback  ProxyCallbackFn
	resultSeq core.Seq
}

func (p *Proxy[T]) LocalAddr() net.Addr {
	return p.proxy.LocalAddr()
}

func (p *Proxy[T]) RemoteAddr() net.Addr {
	return p.proxy.RemoteAddr()
}

func (p *Proxy[T]) At() time.Time {
	return p.proxy.At()
}

func (p *Proxy[T]) Seq() core.Seq {
	return p.proxy.Seq()
}

func (p *Proxy[T]) Send(result core.Result) (n int, err error) {
	n, err = p.proxy.Send(result)
	if err != nil {
		return
	}
	p.resultSeq += 1
	p.callback(hooks.ReceivedResult{Seq: p.resultSeq, Size: n, Result: result})
	return
}

func (p *Proxy[T]) SendWithDeadline(deadline time.Time, result core.Result) (
	n int, err error,
) {
	n, err = p.proxy.SendWithDeadline(deadline, result)
	if err != nil {
		return
	}
	p.resultSeq += 1
	p.callback(hooks.ReceivedResult{Seq: p.resultSeq, Size: n, Result: result})
	return
}
