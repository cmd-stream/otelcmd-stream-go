# otelcmd-stream-go

[![Go Reference](https://pkg.go.dev/badge/github.com/cmd-stream/otelcmd-stream-go.svg)](https://pkg.go.dev/github.com/cmd-stream/otelcmd-stream-go)
[![GoReportCard](https://goreportcard.com/badge/cmd-stream/otelcmd-stream-go)](https://goreportcard.com/report/github.com/cmd-stream/otelcmd-stream-go)
[![codecov](https://codecov.io/gh/cmd-stream/otelcmd-stream-go/graph/badge.svg?token=04UEO65CLJ)](https://codecov.io/gh/cmd-stream/otelcmd-stream-go)

`otelcmd-stream` provides comprehensive OpenTelemetry support for [cmd-stream](https://github.com/cmd-stream/cmd-stream-go), enabling seamless observability for your command-stream based applications.

## Features

- **Distributed Tracing**: Automatic span creation for command sending and execution.
- **Metrics**: Built-in reporting for command execution counts, durations, and status.
- **Context Propagation**: Trace context is automatically carried across the network from sender to server.
- **Customizable**: Hooks for adding business-specific attributes and events to spans and metrics.

## Installation

```bash
go get github.com/cmd-stream/otelcmd-stream-go
```

## Table of Contents

- [Getting Started](#getting-started)
  - [Sender Instrumentation](#sender-instrumentation)
  - [Server Instrumentation](#server-instrumentation)
  - [Traceable Commands](#traceable-commands)

To integrate `otelcmd-stream` into your application, follow these steps:

1. Instrument the **Sender**.
2. Instrument the **Server**.
3. Use **Traceable Commands** to propagate span context.

Use `otelcmd.NewHooksFactory` to instrument your sender:

```go
import (
  "net"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
  sndr "github.com/cmd-stream/cmd-stream-go/sender"
  grp "github.com/cmd-stream/cmd-stream-go/group"
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
    sndr.WithClientsCount[T](clientsCount),
    sndr.WithGroup[T](grp.WithReconnect[T]()),
    sndr.WithSender[T](sndr.WithHooksFactory[T](hooksFactory)),
  )
)
```

Or combine it with a circuit breaker:

```go
import (
  "github.com/ymz-ncnk/circbrk-go"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
  sndr "github.com/cmd-stream/cmd-stream-go/sender"
  hks "github.com/cmd-stream/cmd-stream-go/sender/hooks"
)

var (
  // Create a circuit breaker.
  cb = circbrk.New(  
    circbrk.WithWindowSize(...),
    circbrk.WithFailureRate(...),
    circbrk.WithOpenDuration(...),
    circbrk.WithSuccessThreshold(...),
  )

  // Create OpenTelemetry hooks factory.
  otelHooksFactory = otelcmd.NewHooksFactory[T](
    otelcmd.WithServerAddr[T](serverAddr))

  // Wrap OpenTelemetry hooks factory.
  hooksFactory = hks.NewCircuitBreakerHooksFactory[T](cb, otelHooksFactory)

  // Create sender.
  sender, err = cmdstream.NewSender[T](serverAddr.String(), codec,
    sndr.WithClientsCount[T](clientsCount),
    sndr.WithSender[T](sndr.WithHooksFactory[T](hooksFactory)),
  )
)
```

### Server Instrumentation

Wrap your existing invoker with `otelcmd.NewInvoker` during the server initialization:

```go
import (
  srv "github.com/cmd-stream/cmd-stream-go/server"
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
  cmdstream "github.com/cmd-stream/cmd-stream-go"
)

var (
  codec = ...
  receiver = ...

  // Initialize your invoker and wrap it with OpenTelemetry support.
  invoker = otelcmd.NewInvoker[T](
    srv.NewInvoker[T](receiver),
    otelcmd.WithServerAddr[T](serverAddr),
    // Other available server options:
    // otelcmd.WithPropagator[T](...),
    // otelcmd.WithTracerProvider[T](...),
    // otelcmd.WithMeterProvider[T](...),
  )
  server, err = cmdstream.NewServerWithInvoker[T](invoker, codec, ...)
)
```

### Traceable Commands

For each Command type, define a corresponding traceable type to enable trace context propagation:

```go
type YourTraceCmd = otelcmd.TraceCmd[YourReceiver, YourCmd]
```

Use this type to generate a serializer for the `core.Cmd` interface. To send a `YourCmd` with span context propagation:

```go
import (
  otelcmd "github.com/cmd-stream/otelcmd-stream-go"
)

var (
  ctx = ...
  cmd = otelcmd.NewTraceCmd[YourReceiver, YourCmd](YourCmd{})
)
result, err := sender.Send(ctx, cmd)
```

A full working example is available [here](https://github.com/cmd-stream/examples-go/tree/main/otel).
