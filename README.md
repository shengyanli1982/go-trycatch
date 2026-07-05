<p align="center">
  <img src="assets/logo.png" alt="go-trycatch logo" width="350px">
</p>

<h1 align="center">go-trycatch</h1>

<p align="center"><strong>Try-catch-finally for Go, done right.</strong></p>

<p align="center">
  <a href="https://goreportcard.com/report/github.com/shengyanli1982/go-trycatch"><img alt="Go Report Card" src="https://goreportcard.com/badge/github.com/shengyanli1982/go-trycatch"></a>
  <a href="https://github.com/shengyanli1982/go-trycatch/actions/workflows/test.yaml"><img alt="Build Status" src="https://github.com/shengyanli1982/go-trycatch/actions/workflows/test.yaml/badge.svg"></a>
  <a href="https://pkg.go.dev/github.com/shengyanli1982/go-trycatch"><img alt="Go Reference" src="https://pkg.go.dev/badge/github.com/shengyanli1982/go-trycatch.svg"></a>
  <a href="https://deepwiki.com/shengyanli1982/law"><img alt="Ask DeepWiki" src="https://deepwiki.com/badge.svg"></a>
</p>

go-trycatch adds the familiar try-catch-finally pattern to Go. Panics are recovered and converted to errors, `finally` always runs, and `Do()` returns a proper Go error. No exceptions hierarchy, no surprises, zero dependencies.

## Install

```bash
go get github.com/shengyanli1982/go-trycatch
```

## Quick Start

```go
package main

import (
	"errors"
	"fmt"

	gtc "github.com/shengyanli1982/go-trycatch"
)

func main() {
	err := gtc.New().
		Try(func() error {
			return errors.New("something went wrong")
		}).
		Catch(func(err error) {
			fmt.Printf("caught: %v\n", err)
		}).
		Finally(func() {
			fmt.Println("cleanup")
		}).
		Do()

	fmt.Printf("returned: %v\n", err)
}
// caught: something went wrong
// cleanup
// returned: something went wrong
```

## Features

| Feature             | Description                                                        |
| ------------------- | ------------------------------------------------------------------ |
| Chainable API       | `Try`, `Catch`, `Finally` compose fluently                         |
| Panic recovery      | Panics in try/catch are captured and converted to `error`          |
| `finally` guarantee | Always executes — even when catch panics                           |
| Generic return      | `TryWithResult[T]` and `TryCatchR[T]` for typed results            |
| Context-aware       | `TryCtx` + `WithContext(ctx)` for cancellation and timeouts        |
| Hooks               | `OnTryStart`, `OnTryEnd`, `OnCatch`, `OnFinally` for observability |
| Object pooling      | `Reset()` + `sync.Pool` for zero-allocation reuse                  |
| Zero dependencies   | Standard library only                                              |

## Core API

```go
// Constructor
func New() *TryCatchBlock
func NewWithOptions(opts ...Option) *TryCatchBlock

// Chainable methods
func (tc *TryCatchBlock) Try(fn func() error) *TryCatchBlock
func (tc *TryCatchBlock) TryCtx(fn func(context.Context) error) *TryCatchBlock
func (tc *TryCatchBlock) Catch(fn func(error)) *TryCatchBlock
func (tc *TryCatchBlock) Finally(fn func()) *TryCatchBlock
func (tc *TryCatchBlock) ApplyOptions(opts ...Option) *TryCatchBlock
func (tc *TryCatchBlock) Reset()

// Execute
func (tc *TryCatchBlock) Do() error
```

### Options

| Option             | Description                       |
| ------------------ | --------------------------------- |
| `WithContext(ctx)` | Adds cancellation/timeout support |
| `WithHooks(hooks)` | Registers observability callbacks |
| `WithName(name)`   | Assigns an identifier             |

```go
type Hooks struct {
    OnTryStart func()
    OnTryEnd   func(error)
    OnCatch    func(error)
    OnFinally  func()
}
```

## Generic Functions

For operations that return a typed value, use the generic helpers instead of `Do()`:

```go
// Panic-only recovery
result, err := gtc.TryWithResult(func() (int, error) {
    return 42, nil
})

// With finally cleanup
result, err := gtc.TryWithResultAndFinally(
    func() (string, error) { return "hello", nil },
    func() { fmt.Println("cleanup") },
)

// Full try-catch-finally with typed result
result, err := gtc.TryCatchR(
    func() (int, error) { return computeResult() },
    func(err error) { log.Printf("error: %v", err) },
    func() { fmt.Println("always runs") },
)
```

## Usage Patterns

### Panic Recovery

```go
gtc.New().
    Try(func() error {
        panic("unexpected error")
    }).
    Catch(func(err error) {
        fmt.Printf("caught panic: %v\n", err) // caught panic: unexpected error
    }).
    Do()
```

### Context Cancellation

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

err := gtc.New().
    ApplyOptions(gtc.WithContext(ctx)).
    TryCtx(func(ctx context.Context) error {
        // ctx is passed directly into your function
        return doSomething(ctx)
    }).
    Finally(func() { /* always runs */ }).
    Do()
// If ctx is cancelled before execution, returns context.Canceled
```

### Observability with Hooks

```go
gtc.New().
    ApplyOptions(gtc.WithHooks(gtc.Hooks{
        OnTryStart: func()                  { log.Println("try start") },
        OnTryEnd:   func(err error)         { log.Printf("try end: %v", err) },
        OnCatch:    func(err error)         { log.Printf("caught: %v", err) },
        OnFinally:  func()                  { log.Println("finally") },
    })).
    Try(func() error { return riskyOp() }).
    Do()
```

### Object Pooling (Zero-alloc Reuse)

```go
pool := sync.Pool{New: func() any { return gtc.New() }}

// In a goroutine:
tc := pool.Get().(*gtc.TryCatchBlock)
tc.Try(func() error { return process() }).Do()
tc.Reset()
pool.Put(tc)
```

`Reset()` clears all fields (try, catch, finally, context, hooks, name), making the instance safe to reuse. Benchmarks confirm **zero extra allocations** when using pool mode.

## Execution Flow

```text
Do()
    ├─ context cancelled? ─── return ctx.Err(), finally
    ├─ OnTryStart()
    ├─ try() ── returns error ── OnTryEnd(err) ── OnCatch(err) ── catch(err) ── OnFinally() ── finally()
    └─ try() ── panic ──────── recover() ──────────────────────── OnCatch(err) ── catch(err) ── OnFinally() ── finally()
                                                                                   └─ catch panic? ── re-panic after finally
```

**Key guarantee**: `finally` always executes exactly once, regardless of success, error, or panic paths. If `catch` itself panics, `finally` still runs before the panic propagates.

## Performance

Benchmark results on 12th Gen Intel i5-12400F:

| Path                                | ns/op | allocs/op |
| ----------------------------------- | ----: | --------: |
| `Do()` no error (hot path)          |    ~8 |         0 |
| `Do()` error + catch                |   ~28 |         1 |
| `Do()` error + catch + finally      |   ~29 |         1 |
| `Do()` panic                        |  ~113 |         1 |
| `TryWithResult` no error            |    ~4 |         0 |
| `TryCatchR` error + catch + finally |   ~10 |         0 |
| `Pool` reuse (Get + Reset + Put)    |   ~17 |         0 |

The hot (no-error) path allocates nothing. `Pool` mode is recommended for high-throughput scenarios — it eliminates all per-call allocations.

## Examples

- [Chain call](./examples/chain_call)
- [Concurrent with pool](./examples/concurrent_with_pool)

## Limitations

- Not a replacement for `if err != nil` — a complement for cases where you need catch-finally semantics.
- Error types are not matched by the library. Use `errors.Is` / `errors.As` inside your `Catch` handler.
- Not goroutine-safe by design. One `TryCatchBlock` per goroutine (use `sync.Pool` for sharing).

## License

[MIT](./LICENSE)
