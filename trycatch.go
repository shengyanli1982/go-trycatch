package gotrycatch

import (
	"fmt"
)

// TryCatchBlock implements try-catch-finally error handling pattern
// TryCatchBlock 实现类似于 try-catch-finally 的错误处理模式
type TryCatchBlock struct {
	try     func() error // Function to execute that may return an error
	catch   func(error)  // Function to handle any errors from try
	finally func()       // Function that always executes after try-catch
}

// New returns a TryCatchBlock instance
// New 返回一个 TryCatchBlock 实例
func New() *TryCatchBlock {
	return &TryCatchBlock{}
}

// Reset cleans up the block state (useful for object pooling)
// Reset 清理块的状态 (方便作为对象池复用)
func (tc *TryCatchBlock) Reset() {
	tc.try = nil
	tc.catch = nil
	tc.finally = nil
}

// Try sets the main execution function
// Try 设置主要执行函数，该函数可能产生错误
func (tc *TryCatchBlock) Try(try func() error) *TryCatchBlock {
	tc.try = try
	return tc
}

// Catch sets the error handling function
// Catch 设置错误处理函数，用于处理来自 Try 的错误
func (tc *TryCatchBlock) Catch(catch func(error)) *TryCatchBlock {
	tc.catch = catch
	return tc
}

// Finally sets the cleanup function that always executes
// Finally 设置清理函数，该函数会在所有情况下都被执行
func (tc *TryCatchBlock) Finally(finally func()) *TryCatchBlock {
	tc.finally = finally
	return tc
}

// Do executes the try-catch-finally block in sequence
// Do 按顺序执行 try-catch-finally 流程，包括错误处理和 panic 恢复
func (tc *TryCatchBlock) Do() {
	// Validate try function exists
	// 验证 try 函数是否存在
	if tc.try == nil {
		return
	}

	// Recover from panics and convert them to errors
	// 从 panic 中恢复并将其转换为标准错误
	defer func() {
		// Handle panic first
		// 1. 首先处理 panic（如果有的话）
		if r := recover(); r != nil {
			var err error
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
			if tc.catch != nil {
				tc.catch(err)
			}
		}

		// Execute finally if it exists
		// 2. 执行 finally（如果有的话）
		if tc.finally != nil {
			tc.finally()
		}

		// Reset the block
		// 3. 最后执行 Reset
		tc.Reset()
	}()

	// Execute try and handle any returned errors
	// 执行 try 函数并处理返回的错误
	if err := tc.try(); err != nil && tc.catch != nil {
		tc.catch(err)
	}
}
