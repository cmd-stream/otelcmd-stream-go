package otelcmd

import (
	"net"
	"time"

	"github.com/cmd-stream/core-go"
	"github.com/cmd-stream/sender-go/hooks"
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

func (p *Proxy[T]) Send(seq core.Seq, result core.Result) (n int, err error) {
	n, err = p.proxy.Send(seq, result)
	if err != nil {
		return
	}
	p.resultSeq += 1
	p.callback(hooks.ReceivedResult{Seq: p.resultSeq, Size: n, Result: result})
	return
}

func (p *Proxy[T]) SendWithDeadline(seq core.Seq, result core.Result,
	deadline time.Time) (n int, err error) {
	n, err = p.proxy.SendWithDeadline(seq, result, deadline)
	if err != nil {
		return
	}
	p.resultSeq += 1
	p.callback(hooks.ReceivedResult{Seq: p.resultSeq, Size: n, Result: result})
	return
}
