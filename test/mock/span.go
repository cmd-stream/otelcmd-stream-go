package mock

import (
	"github.com/ymz-ncnk/mok"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type EndFn func(options ...trace.SpanEndOption)
type AddEventFn func(name string, options ...trace.EventOption)
type AddLinkFn func(link trace.Link)
type IsRecordingFn func() bool
type RecordErrorFn func(err error, options ...trace.EventOption)
type SpanContextFn func() trace.SpanContext
type SetStatusFn func(code codes.Code, description string)
type SetNameFn func(name string)
type SetAttributesFn func(attrs ...attribute.KeyValue)
type TraceContextFn func() trace.TracerProvider

func NewSpan() Span {
	return Span{Mock: mok.New("Span")}
}

type Span struct {
	trace.Span
	*mok.Mock
}

func (s Span) RegisterEnd(fn EndFn) Span {
	s.Register("End", fn)
	return s
}

func (s Span) RegisterAddEvent(fn AddEventFn) Span {
	s.Register("AddEvent", fn)
	return s
}

func (s Span) RegisterAddLink(fn AddLinkFn) Span {
	s.Register("AddLink", fn)
	return s
}

func (s Span) RegisterIsRecording(fn IsRecordingFn) Span {
	s.Register("IsRecording", fn)
	return s
}

func (s Span) RegisterRecordError(fn RecordErrorFn) Span {
	s.Register("RecordError", fn)
	return s
}

func (s Span) RegisterSpanContext(fn SpanContextFn) Span {
	s.Register("SpanContext", fn)
	return s
}

func (s Span) RegisterSetStatus(fn SetStatusFn) Span {
	s.Register("SetStatus", fn)
	return s
}

func (s Span) RegisterSetName(fn SetNameFn) Span {
	s.Register("SetName", fn)
	return s
}

func (s Span) RegisterSetAttributes(fn SetAttributesFn) Span {
	s.Register("SetAttributes", fn)
	return s
}

func (s Span) RegisterTracerProvider(fn TraceContextFn) Span {
	s.Register("TracerProvider", fn)
	return s
}

func (s Span) End(options ...trace.SpanEndOption) {
	_, err := s.Call("End", options)
	if err != nil {
		panic(err)
	}
}

func (s Span) AddEvent(name string, options ...trace.EventOption) {
	_, err := s.Call("AddEvent", name, options)
	if err != nil {
		panic(err)
	}
}

func (s Span) AddLink(link trace.Link) {
	_, err := s.Call("AddLink", link)
	if err != nil {
		panic(err)
	}
}

func (s Span) IsRecording() bool {
	results, err := s.Call("IsRecording")
	if err != nil {
		panic(err)
	}
	return results[0].(bool)
}

func (s Span) RecordError(err error, options ...trace.EventOption) {
	_, aerr := s.Call("RecordError", err, options)
	if aerr != nil {
		panic(aerr)
	}
}

func (s Span) SpanContext() trace.SpanContext {
	results, err := s.Call("SpanContext")
	if err != nil {
		panic(err)
	}
	return results[0].(trace.SpanContext)
}

func (s Span) SetStatus(code codes.Code, description string) {
	_, err := s.Call("SetStatus", code, description)
	if err != nil {
		panic(err)
	}
}

func (s Span) SetName(name string) {
	_, err := s.Call("SetName", name)
	if err != nil {
		panic(err)
	}
}

func (s Span) SetAttributes(kv ...attribute.KeyValue) {
	_, err := s.Call("SetAttributes", kv)
	if err != nil {
		panic(err)
	}
}

func (s Span) TracerProvider() trace.TracerProvider {
	results, err := s.Call("TracerProvider")
	if err != nil {
		panic(err)
	}
	tp, _ := results[0].(trace.TracerProvider)
	return tp
}
