[English](./README.md) | 中文

<div align="center">
	<h1>go-trycatch</h1>
    <p>一个优雅简洁的 Go 语言 try-catch-finally 错误处理实现</p>
	<img src="assets/logo.png" alt="logo" width="350px">
</div>

[![Go Report Card](https://goreportcard.com/badge/github.com/shengyanli1982/go-trycatch)](https://goreportcard.com/report/github.com/shengyanli1982/go-trycatch)
[![Build Status](https://github.com/shengyanli1982/go-trycatch/actions/workflows/test.yaml/badge.svg)](github.com/shengyanli1982/go-trycatch/actions)
[![Go Reference](https://pkg.go.dev/badge/github.com/shengyanli1982/go-trycatch.svg)](https://pkg.go.dev/github.com/shengyanli1982/go-trycatch)

# 简介

一直想在 Go 中使用 try-catch 代码块？现在这个愿望实现了！`go-trycatch` 为 Go 带来了熟悉的 try-catch-finally 模式。它的独特之处在于：我们的目标是对 Go 现有的错误处理机制进行补充，而不是替代。你可以把它想象成是为你的错误处理增添一份超能力，同时又完美保持了 Go 代码的地道风格。🦸‍♂️

`go-trycatch` 的核心特性：

1. 符合直觉的 Try-catch-finally 模式（无需复杂的异常层级）
2. 智能转换 panic 为可控的错误类型
3. 优雅流畅的链式调用 API
4. 纯净无依赖（极简主义设计）
5. 可靠的 Finally 执行保证
6. 与 Go 生态完美融合（零侵入性）

# 为什么选择 go-trycatch？

-   **简洁优雅的 API**：直观的语法设计，让错误处理不再繁琐
-   **全面的错误管理**：通过熟悉的 try-catch 模式轻松处理各类错误
-   **可靠的 Panic 处理**：自动将 panic 转化为标准错误
-   **资源释放保证**：Finally 块确保资源正确清理
-   **轻量级设计**：零依赖，最小化项目负担
-   **链式调用体验**：流畅的方法链接口，提升代码可读性
-   **原生兼容性**：与 Go 标准错误处理无缝协作
-   **并发安全性**：不保证 goroutine 安全性，需在并发场景中自行确保安全

# 安装

一行命令即可安装：

```bash
go get github.com/shengyanli1982/go-trycatch
```

# 快速开始

来看一个简单示例 - 简单到不能再简单：

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

**运行效果**

```bash
$ go run demo.go
Caught error: something went wrong
Cleanup code here
```

# 功能特性

## 1. 三大核心：Try、Catch 和 Finally

就像一个完美的故事需要三个主角：

-   `Try`：放置主要业务逻辑的地方
-   `Catch`：优雅处理错误的场所
-   `Finally`：确保资源清理的保障

### 使用示例

```go
New().
    Try(func() error {
        // 核心业务逻辑
        return someFunction()
    }).
    Catch(func(err error) {
        // 错误处理
        log.Printf("捕获到错误：%v", err)
    }).
    Finally(func() {
        // 资源清理
        closeResources()
    }).
    Do()
```

## 2. Panic 处理

`go-trycatch` 能自动捕获 panic 并将其转换为标准错误，让你在 catch 块中统一处理。

### Panic 处理示例

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

## 3. Finally 保证

无论代码执行是否出错，Finally 块都能确保资源得到妥善处理。

### 资源清理示例

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

# 使用限制

为了保持透明，我们需要说明 `go-trycatch` 的一些限制：

-   不是 Go 原生错误处理的替代品（我们的目标是锦上添花，而非取而代之）
-   不是性能至上的解决方案（为了便利性，会有些许性能开销）
-   不能直接捕获指定的错误类型（但可以在 Catch 函数中自行实现）

    ```go
    // 示例：在 Catch 中处理特定错误类型
    var ErrNotFound = errors.New("not found")

    New().
        Try(func() error {
            return ErrNotFound
        }).
        Catch(func(err error) {
            // 在 Catch 中手动判断错误类型
            if errors.Is(err, ErrNotFound) {
                fmt.Println("处理未找到错误")
            } else {
                fmt.Println("处理其他错误")
            }
        }).
        Do()
    ```

-   不提供内置的错误类型匹配功能（需要在 Catch 函数中手动处理）

    ```go
    // 示例：在 Catch 中处理多种错误类型
    type CustomError struct {
        Code    int
        Message string
    }

    func (e *CustomError) Error() string {
        return e.Message
    }

    New().
        Try(func() error {
            return &CustomError{Code: 404, Message: "资源未找到"}
        }).
        Catch(func(err error) {
            // 手动进行类型断言和错误处理
            if customErr, ok := err.(*CustomError); ok {
                switch customErr.Code {
                case 404:
                    fmt.Println("处理 404 错误:", customErr.Message)
                case 500:
                    fmt.Println("处理 500 错误:", customErr.Message)
                }
            }
        }).
        Do()
    ```

# 参与贡献

我们欢迎任何形式的贡献！如果你有好的想法或建议，请随时提交 Pull Request。

# 开源协议

`go-trycatch` 采用 MIT 开源协议。详细信息请查看 LICENSE 文件。
