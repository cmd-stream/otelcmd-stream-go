# otelcmd-stream: OpenTelemetry for cmd-stream

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/otelcmd-stream-go.svg)](https://pkg.go.dev/github.com/cmd-stream/otelcmd-stream-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/otelcmd-stream-go)](https://goreportcard.com/report/github.com/cmd-stream/otelcmd-stream-go)
[![codecov](https://codecov.io/gh/cmd-stream/otelcmd-stream-go/graph/badge.svg?token=04UEO65CLJ)](https://codecov.io/gh/cmd-stream/otelcmd-stream-go)

## Table of Contents

- [otelcmd-stream: OpenTelemetry for cmd-stream](#otelcmd-stream-opentelemetry-for-cmd-stream)
  - [Table of Contents](#table-of-contents)
  - [Features](#features)
  - [Installation](#installation)
  - [Usage](#usage)
    - [Sender](#sender)
    - [Server Instrumentation](#server-instrumentation)
    - [Traceable Commands](#traceable-commands)

## Features

- Automated Tracing: Automatically creates spans for Command transmission and
  execution.
- Built-in Metrics: Reports execution counts, durations, and success/failure
  statuses out of the box.
- Context Propagation: Transparently carries trace context across the network 
  via the `TraceCmd` wrapper.
- Extensible Hooks: Inject business-specific attributes and events into spans 
  and metrics.

## Installation

```bash
go get github.com/cmd-stream/otelcmd-stream-go
```

## Usage

1. Instrument the sender.
2. Instrument the server.
3. Use `TraceCmd` for context propagation.

### Sender

Instrument the sender with `otelcmd.NewHooksFactory`:

```go
import (
  "net"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
  sndr "github.com/cmd-stream/cmd-stream-go/sender"
)

var (
  serverAddr net.Addr = ...
  codec = ...

  // Create OpenTelemetry hooks factory with optional customization.
  hooksFactory = otelcmd.NewHooksFactory[T](
    otelcmd.WithServerAddr[T](serverAddr),
    // Other available options:
    // otelcmd.WithPropagator[T](...),
    // otelcmd.WithTracerProvider[T](...),
    // otelcmd.WithMeterProvider[T](...),
    // otelcmd.WithSpanAttributesFn[T](...),
  )

  // Initialize the high-level sender with instrumentation.
  sender, err = cmdstream.NewSender[T](serverAddr.String(), codec,
    sndr.WithSender[T](sndr.WithHooksFactory[T](hooksFactory)),
  )
)
```

You can also use a Circuit Breaker:

```go
import (
  "github.com/ymz-ncnk/circbrk-go"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
  sndr "github.com/cmd-stream/cmd-stream-go/sender"
  hks "github.com/cmd-stream/cmd-stream-go/sender/hooks"
)

var (
  // 1. Create a circuit breaker.
  cb = circbrk.New(
    circbrk.WithWindowSize(...),
    circbrk.WithFailureRate(...),
    circbrk.WithOpenDuration(...),
    circbrk.WithSuccessThreshold(...),
  )

  // 2. Create the OTel hooks factory.
  otelHooksFactory = otelcmd.NewHooksFactory[T](
    otelcmd.WithServerAddr[T](serverAddr))

  // 3. Compose them together.
  hooksFactory = hks.NewCircuitBreakerHooksFactory[T](cb, otelHooksFactory)

  // 4. Initialize the sender with the combined hooks.
  sender, err = cmdstream.NewSender[T](serverAddr.String(), codec,
    sndr.WithSender[T](sndr.WithHooksFactory[T](hooksFactory)),
  )
)
```

### Server Instrumentation

Wrap your invoker with `otelcmd.NewInvoker`:

```go
import (
  srv "github.com/cmd-stream/cmd-stream-go/server"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
)

var (
  // Initialize your invoker and wrap it with OpenTelemetry support.
  invoker = otelcmd.NewInvoker[T](
    srv.NewInvoker[T](receiver),
    otelcmd.WithServerAddr[T](serverAddr),
    // Other available options:
    // otelcmd.WithPropagator[T](...),
    // otelcmd.WithTracerProvider[T](...),
    // otelcmd.WithMeterProvider[T](...),
  )
  server, err = cmdstream.NewServerWithInvoker[T](invoker, codec, ...)
)
```

### Traceable Commands

`TraceCmd` carries the span context across the network. Define a traceable type 
which is used to generate the `core.Cmd` serializer:

```go
type YourTraceCmd = otelcmd.TraceCmd[YourReceiver, YourCmd]
```

Use `otelcmd.NewTraceCmd` to wrap your Commands:

```go
import (
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
)

var (
  ctx = ...
  cmd YourTraceCmd = otelcmd.NewTraceCmd[YourReceiver, YourCmd](YourCmd{})
)
result, err := sender.Send(ctx, cmd)
```

A full working example is available [here](https://github.com/cmd-stream/examples-go/tree/main/otel).
