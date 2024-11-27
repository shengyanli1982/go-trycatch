package gotrycatch

import (
	"errors"
	"fmt"
)

var (
	// ErrorTryFuncNil is returned when the try function is nil
	// ErrorTryFuncNil 在 try 函数为空时返回
	ErrorTryFuncNil = errors.New("try function is nil")
)

// TryCatchBlock defines an error handling block
// TryCatchBlock 定义一个错误处理块，用于实现类似 try-catch 的错误处理机制
type TryCatchBlock struct {
	try     func() error // Main execution function 主要执行函数
	catch   func(error)  // Error handling function 错误处理函数
	finally func()       // Cleanup function 清理函数
}

// New creates a new error handling block
// New 创建并返回一个新的错误处理块实例
func New() *TryCatchBlock {
	return &TryCatchBlock{}
}

// reset resets the error handling block
// reset 重置错误处理块
func (tc *TryCatchBlock) reset() {
	tc.try = nil
	tc.catch = nil
	tc.finally = nil
}

// Try adds the execution function to the block
// Try 添加要执行的主函数，该函数可能返回错误
func (tc *TryCatchBlock) Try(try func() error) *TryCatchBlock {
	tc.try = try
	return tc
}

// Catch adds the error handling function to the block
// Catch 添加错误处理函数，用于处理 try 中发生的错误
func (tc *TryCatchBlock) Catch(catch func(error)) *TryCatchBlock {
	tc.catch = catch
	return tc
}

// Finally adds the cleanup function to the block
// Finally 添加最终清理函数，该函数总是会被执行
func (tc *TryCatchBlock) Finally(finally func()) *TryCatchBlock {
	tc.finally = finally
	return tc
}

// Do executes the error handling block
// Do 按顺序执行整个错误处理块：try、catch（如果发生错误）和 finally
func (tc *TryCatchBlock) Do() (err error) {
	// Ensure try function is not nil
	// 确保 try 函数不为空
	if tc.try == nil {
		return ErrorTryFuncNil
	}

	// Reset the error handling block
	// 重置错误处理块
	defer tc.reset()

	// Execute finally function before function returns
	// 确保在函数返回前执行 finally 函数
	if tc.finally != nil {
		defer func() { tc.finally() }()
	}

	// Handle panic and convert it to error
	// 处理 panic 并将其转换为 error
	defer func() {
		if r := recover(); r != nil {
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
	}()

	// Execute try function and handle any errors
	// 执行 try 函数并处理可能发生的错误
	err = tc.try()
	if err != nil && tc.catch != nil {
		tc.catch(err)
	}

	return err
}
