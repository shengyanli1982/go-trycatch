English | [中文](./README_CN.md)

<div align="center">
	<h1>go-trycatch</h1>
    <p>A simple, elegant implementation of try-catch-finally error handling pattern for Go.</p>
	<img src="assets/logo.png" alt="logo" width="350px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/go-trycatch)](https://goreportcard.com/report/github.com/shengyanli1982/go-trycatch)
[![Build Status](https://github.com/shengyanli1982/go-trycatch/actions/workflows/test.yaml/badge.svg)](github.com/shengyanli1982/go-trycatch/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/go-trycatch.svg)](https://pkg.go.dev/github.com/shengyanli1982/go-trycatch)

# Introduction

Ever wished Go had try-catch blocks? Well, wish granted! `go-trycatch` brings the familiar try-catch-finally pattern to Go, but with a twist - it's designed to complement Go's existing error handling philosophy, not replace it. Think of it as giving your error handling superpowers while keeping your Go code idiomatic. 🦸‍♂️

`go-trycatch` comes packed with goodies:

1. Try-catch-finally pattern that feels like home (minus the endless exception hierarchies)
2. Automatic panic recovery that turns those scary panics into manageable errors
3. A chainable API so smooth it feels like writing poetry
4. Zero external dependencies (we're not that kind of library)
5. Finally blocks that actually run... finally
6. Plays nicely with your existing Go code (no drama)

# Why Choose go-trycatch?

-   **Intuitive API Design**: Clean, expressive syntax that makes error handling a breeze
-   **Flexible Error Management**: Comprehensive error handling with familiar try-catch patterns
-   **Robust Panic Recovery**: Automatically converts panic situations into manageable errors
-   **Guaranteed Cleanup**: Ensures resource cleanup with reliable finally block execution
-   **Lightweight Integration**: Zero external dependencies for minimal project overhead
-   **Fluent Method Chaining**: Enables clear, readable code flow through chainable operations
-   **Native Go Compatibility**: Seamlessly integrates with Go's existing error handling patterns
-   **Concurrency Safety**: Does not guarantee goroutine safety, ensure safety in concurrent scenarios yourself

# Installation

To install `go-trycatch`, use the `go get` command:

```bash
go get github.com/shengyanli1982/go-trycatch
```

# Quick Start

Here's a taste of `go-trycatch` in action - it's so simple, your cat could probably use it:

```go
package main

import (
    "fmt"
    gtc "github.com/shengyanli1982/go-trycatch"
)

func main() {
    gtc.New().
        Try(func() error {
            // Your code that might return error or panic
            return fmt.Errorf("something went wrong")
        }).
        Catch(func(err error) {
            fmt.Printf("Caught error: %v\n", err)
        }).
        Finally(func() {
            fmt.Println("Cleanup code here")
        }).
        Do()
}
```

**Result**

```bash
$ go run demo.go
Caught error: something went wrong
Cleanup code here
```

# Features

## 1. The Three Musketeers: Try, Catch, and Finally

Every good story needs heroes, and here are ours:

-   `Try`: Where your brave code ventures forth
-   `Catch`: Your safety net when things go south
-   `Finally`: The cleanup crew that never calls in sick

### Example

```go
New().
    Try(func() error {
        // Main logic here
        return someFunction()
    }).
    Catch(func(err error) {
        // Error handling logic
        log.Printf("An error occurred: %v", err)
    }).
    Finally(func() {
        // Cleanup logic
        closeResources()
    }).
    Do()
```

## 2. Panic Recovery

`go-trycatch` automatically recovers from panics and converts them to errors that can be handled in the catch block.

### Example with Panic Handling

```go
New().
    Try(func() error {
        panic("unexpected error")
    }).
    Catch(func(err error) {
        fmt.Printf("Caught panic: %v\n", err)
    }).
    Do()
```

## 3. Cleanup with Finally

The `Finally` block ensures proper resource cleanup, executing regardless of whether an error occurred.

### Example with Resource Cleanup

```go
New().
    Try(func() error {
        return useResource(resource)
    }).
    Catch(func(err error) {
        log.Printf("Error using resource: %v", err)
    }).
    Finally(func() {
        releaseResource(resource)
    }).
    Do()
```

# Best Practices

## 1. Instance Reusability

The `TryCatchBlock` instance is designed to be reusable. When reusing an instance, always call the `Reset()` method between operations to ensure proper state management.

```go
// Initialize a reusable instance
tryCatch := New().
    Try(func() error {
        return processResource(resource)
    }).
    Catch(func(err error) {
        log.Printf("Resource processing error: %v", err)
    }).
    Finally(func() {
        cleanupResource(resource)
    }).
    Do()

// Reset the instance state
tryCatch.Reset()

// Reuse the instance for another operation
tryCatch.Try(func() error {
        return processSecondaryResource(secondaryResource)
    }).
    Catch(func(err error) {
        log.Printf("Secondary resource processing error: %v", err)
    }).
    Finally(func() {
        cleanupSecondaryResource(secondaryResource)
    }).
    Do()
```

## 2. Object Pooling with `sync.Pool`

For high-throughput scenarios where `TryCatchBlock` instances are frequently created and destroyed, utilize `sync.Pool` to optimize memory allocation and reduce GC pressure.

```go
// Initialize a thread-safe pool of TryCatchBlock instances
var tryCatchPool = sync.Pool{
    New: func() interface{} {
        return New()
    },
}

func handleOperation() {
    // Acquire an instance from the pool
    tryCatch := tryCatchPool.Get().(*TryCatchBlock)

    // Execute the operation
    tryCatch.Try(func() error {
        return processResource(resource)
    }).
    Catch(func(err error) {
        log.Printf("Operation failed: %v", err)
    }).
    Finally(func() {
        cleanupResource(resource)
    }).
    Do()

    // Reset the instance state
    tryCatch.Reset()

    // Return the instance to the pool
    tryCatchPool.Put(tryCatch)
}

// Concurrently handle multiple operations
var wg sync.WaitGroup
for i := 0; i < 1000; i++ {
    wg.Add(1)
    go handleOperation()
}

wg.Wait()
```

# Limitations

Let's be honest about what `go-trycatch` isn't:

-   A complete replacement for Go's error handling (we're here to complement, not conquer)
-   The fastest thing in the world (there's a tiny performance cost for all this convenience)
-   A magic wand for catching specific error types (though you can still do it with a bit of elbow grease)

    ```go
    // eg: handling specific error types in Catch
    var ErrNotFound = errors.New("not found")

    New().
        Try(func() error {
            return ErrNotFound
        }).
        Catch(func(err error) {
            // Manually check error type in Catch
            if errors.Is(err, ErrNotFound) {
                fmt.Println("Handling not found error")
            } else {
                fmt.Println("Handling other errors")
            }
        }).
        Do()
    ```

-   No built-in support for catching specific error types (but you can implement this in your Catch function)

    ```go
    // eg: handling multiple error types in Catch
    type CustomError struct {
        Code    int
        Message string
    }

    func (e *CustomError) Error() string {
        return e.Message
    }

    New().
        Try(func() error {
            return &CustomError{Code: 404, Message: "Resource not found"}
        }).
        Catch(func(err error) {
            // Manually handle error types in Catch
            if customErr, ok := err.(*CustomError); ok {
                switch customErr.Code {
                case 404:
                    fmt.Println("Handling 404 error:", customErr.Message)
                case 500:
                    fmt.Println("Handling 500 error:", customErr.Message)
                }
            }
        }).
        Do()
    ```

# Contributing

Contributions to `go-trycatch` are welcome! Please feel free to submit a Pull Request.

# License

`go-trycatch` is released under the MIT License. See the LICENSE file for details.
