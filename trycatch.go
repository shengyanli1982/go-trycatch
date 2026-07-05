package gotrycatch

import (
	"context"
	"fmt"
)

// stringError 是一个轻量级的 error 实现，用于将非 error 类型的 panic 值转换为 error。
// 相比 fmt.Errorf，它避免了额外的分配和 fmt 包的开销。
type stringError struct {
	msg string
}

func (e stringError) Error() string { return e.msg }

// catchGuard 在隔离环境中执行 catch 函数，捕获 catch 内部的 panic 并返回。
// 返回 nil 表示 catch 正常执行完毕；返回非 nil 表示 catch 发生了 panic。
func catchGuard(fn func(error), err error) (panicVal any) {
	defer func() { panicVal = recover() }()
	fn(err)
	return
}

// TryCatchBlock 实现 try-catch-finally 错误处理模式
type TryCatchBlock struct {
	try     func() error              // 待执行的函数，可能返回错误
	tryCtx  func(context.Context) error // 上下文感知的 try 函数，与 try 互斥
	catch   func(error)               // 错误处理函数
	finally func()                    // 清理函数，在所有情况下都会执行
	ctx     context.Context           // 用于取消和超时的上下文
	hooks   Hooks                     // 监控执行的钩子
	name    string                    // 块的名称标识符
}

// New 返回一个 TryCatchBlock 实例
func New() *TryCatchBlock {
	return &TryCatchBlock{}
}

// Reset 清理块的状态，用于对象池复用
// 注意：Reset 只清理函数指针。如果闭包中捕获了敏感数据，需由调用方确保不会泄露
func (tc *TryCatchBlock) Reset() {
	tc.try = nil
	tc.tryCtx = nil
	tc.catch = nil
	tc.finally = nil
	tc.ctx = nil
	tc.hooks = Hooks{}
	tc.name = ""
}

// Try 设置待执行的函数
func (tc *TryCatchBlock) Try(try func() error) *TryCatchBlock {
	tc.try = try
	return tc
}

// TryCtx 设置上下文感知的 try 函数，与 Try 互斥
func (tc *TryCatchBlock) TryCtx(try func(context.Context) error) *TryCatchBlock {
	tc.tryCtx = try
	return tc
}

// Catch 设置错误处理函数
func (tc *TryCatchBlock) Catch(catch func(error)) *TryCatchBlock {
	tc.catch = catch
	return tc
}

// Finally 设置清理函数
func (tc *TryCatchBlock) Finally(finally func()) *TryCatchBlock {
	tc.finally = finally
	return tc
}

// Do 执行 try-catch-finally 流程，返回错误
// 返回 try 返回的错误或 panic 转换的错误
func (tc *TryCatchBlock) Do() (err error) {
	var (
		ctxCancelled  bool
		catchCalled   bool
		catchPanicErr any
		returnedErr   error
	)

	defer func() {
		// recover() 必须在 defer 函数的顶层调用（不能在内层闭包中调用）
		r := recover()

		// 1. 处理 panic
		if r != nil {
			var panicErr error
			switch v := r.(type) {
			case error:
				panicErr = v
			case string:
				panicErr = stringError{v}
			default:
				panicErr = stringError{fmt.Sprintf("%v", v)}
			}
			if tc.hooks.OnCatch != nil {
				tc.hooks.OnCatch(panicErr)
			}
			if tc.catch != nil && !catchCalled {
				catchPanicErr = catchGuard(tc.catch, panicErr)
			}
			returnedErr = panicErr
			err = panicErr
		} else if !ctxCancelled {
			// 2. 正常路径：处理 try() 返回的错误，调用 catch
			if returnedErr != nil && tc.catch != nil {
				catchCalled = true
				if tc.hooks.OnCatch != nil {
					tc.hooks.OnCatch(returnedErr)
				}
				catchPanicErr = catchGuard(tc.catch, returnedErr)
			}
			err = returnedErr
		}

		// finally 始终执行（catch panic 已被 catchGuard 隔离）
		if tc.hooks.OnFinally != nil {
			tc.hooks.OnFinally()
		}
		if tc.finally != nil {
			tc.finally()
		}

		// 如果 catch 产生了 panic，向上传播
		if catchPanicErr != nil {
			panic(catchPanicErr)
		}
	}()

	if tc.try == nil && tc.tryCtx == nil {
		return nil
	}

	// 检查 context 是否已取消（不再 early return，让 defer 统一处理 finally）
	if tc.ctx != nil {
		select {
		case <-tc.ctx.Done():
			ctxCancelled = true
			err = tc.ctx.Err()
			return
		default:
		}
	}

	// 执行 OnTryStart 钩子
	if tc.hooks.OnTryStart != nil {
		tc.hooks.OnTryStart()
	}

	// 执行 try 函数
	if tc.try != nil {
		returnedErr = tc.try()
	} else if tc.tryCtx != nil {
		ctx := tc.ctx
		if ctx == nil {
			ctx = context.Background()
		}
		returnedErr = tc.tryCtx(ctx)
	}

	// 执行 OnTryEnd 钩子
	if tc.hooks.OnTryEnd != nil {
		tc.hooks.OnTryEnd(returnedErr)
	}

	return
}
