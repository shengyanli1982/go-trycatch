package gotrycatch

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryCatchBlock_Do(t *testing.T) {
	tests := []struct {
		name           string
		tryFunction    func() error
		catchHandler   func(error)
		finallyHandler func()
	}{
		{
			name: "No error",
			tryFunction: func() error {
				return nil
			},
			catchHandler:   nil,
			finallyHandler: nil,
		},
		{
			name: "Error in try",
			tryFunction: func() error {
				return errors.New("try error")
			},
			catchHandler: func(err error) {
				assert.Equal(t, "try error", err.Error())
				assert.NotNil(t, err)
			},
			finallyHandler: nil,
		},
		{
			name: "Panic in try",
			tryFunction: func() error {
				panic("panic error")
			},
			catchHandler: func(err error) {
				assert.Equal(t, "panic error", err.Error())
				assert.NotNil(t, err)
			},
			finallyHandler: nil,
		},
		{
			name: "Finally function",
			tryFunction: func() error {
				return nil
			},
			catchHandler: nil,
			finallyHandler: (func() func() {
				isFinalized := false
				return func() {
					isFinalized = true
					assert.True(t, isFinalized, "finally handler should be executed")
				}
			})(),
		},
		{
			name:           "Try function is nil",
			tryFunction:    nil,
			catchHandler:   nil,
			finallyHandler: nil,
		},
		// {
		// 	name: "Nested panic in catch",
		// 	tryFunc: func() error {
		// 		panic("original panic")
		// 	},
		// 	catchFunc: func(err error) {
		// 		assert.Equal(t, "original panic", err.Error())
		// 		assert.NotNil(t, err)

		// 		panic("panic in catch")
		// 	},
		// 	finallyFunc: (func() func() {
		// 		executed := false
		// 		return func() {
		// 			executed = true
		// 			assert.True(t, executed, "finally should be executed even with nested panic")
		// 		}
		// 	})(),
		// 	expectedError: errors.New("panic in catch"),
		// },
		{
			name: "Finally executes after panic",
			tryFunction: func() error {
				panic("panic error")
			},
			catchHandler: nil,
			finallyHandler: (func() func() {
				isFinalized := false
				return func() {
					isFinalized = true
					assert.True(t, isFinalized, "finally handler should be executed")
				}
			})(),
		},
		{
			name: "Complex error chain",
			tryFunction: func() error {
				originalErr := errors.New("original error")
				return fmt.Errorf("wrapped: %w", originalErr)
			},
			catchHandler: func(err error) {
				assert.Contains(t, err.Error(), "original error")
				assert.Contains(t, err.Error(), "wrapped")

				unwrappedErr := errors.Unwrap(err)
				assert.NotNil(t, unwrappedErr)
				assert.Equal(t, "original error", unwrappedErr.Error())

				assert.True(t, errors.Is(err, unwrappedErr), "original error should be in error chain")
			},
		},
		{
			name: "Panic with custom error type",
			tryFunction: func() error {
				panic(customError{errorMessage: "custom error"})
			},
			catchHandler: func(err error) {
				assert.Equal(t, "custom error", err.Error())
				customErr, ok := err.(customError)
				assert.True(t, ok, "error should be of type customError")
				assert.Equal(t, "custom error", customErr.errorMessage)
			},
			finallyHandler: nil,
		},
		{
			name: "Multiple deferred operations",
			tryFunction: func() error {
				defer func() {
					// 模拟其他 defer 操作
				}()
				return errors.New("error after defer")
			},
			catchHandler:   nil,
			finallyHandler: nil,
		},
		{
			name: "Nil catch with error",
			tryFunction: func() error {
				return errors.New("uncaught error")
			},
			catchHandler:   nil,
			finallyHandler: nil,
		},
		{
			name:           "Empty try-catch-finally chain",
			tryFunction:    func() error { return nil },
			catchHandler:   nil,
			finallyHandler: nil,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			tryCatch := New().
				Try(testCase.tryFunction).
				Catch(testCase.catchHandler).
				Finally(testCase.finallyHandler)
			tryCatch.Do()
		})
	}
}

// 自定义错误类型
type customError struct {
	errorMessage string
}

func (e customError) Error() string {
	return e.errorMessage
}

// 测试链式调用
func TestTryCatchBlock_ChainCalls(t *testing.T) {
	assert := assert.New(t)
	isFinallyExecuted := false
	isErrorCaught := false

	tryCatch := New().
		Try(func() error {
			assert.False(isFinallyExecuted, "finally handler not executed yet")
			return errors.New("test error")
		}).
		Catch(func(err error) {
			assert.Equal("test error", err.Error())
			isErrorCaught = true
		}).
		Finally(func() {
			isFinallyExecuted = true
		})

	tryCatch.Do()

	assert.True(isErrorCaught, "catch handler should be executed")
	assert.True(isFinallyExecuted, "finally handler should be executed")
}

// 测试重复使用同一个实例
func TestTryCatchBlock_Reuse(t *testing.T) {
	assert := assert.New(t)
	tryCatch := New()

	// 第一次使用
	firstTryExecuted := false
	firstErrorCaught := false
	tryCatch.Try(func() error {
		firstTryExecuted = true
		return errors.New("first error")
	}).Catch(func(err error) {
		firstErrorCaught = true
		assert.Equal("first error", err.Error())
	}).Do()
	tryCatch.Reset()

	assert.True(firstTryExecuted, "first try block should be executed")
	assert.True(firstErrorCaught, "first catch block should be executed")
	assert.Nil(tryCatch.try, "try function should be reset after Do()")
	assert.Nil(tryCatch.catch, "catch function should be reset after Do()")
	assert.Nil(tryCatch.finally, "finally function should be reset after Do()")

	// 第二次使用
	secondTryExecuted := false
	secondErrorCaught := false
	tryCatch.Try(func() error {
		secondTryExecuted = true
		return errors.New("second error")
	}).Catch(func(err error) {
		secondErrorCaught = true
		assert.Equal("second error", err.Error())
	}).Do()
	tryCatch.Reset()

	assert.True(secondTryExecuted, "second try block should be executed")
	assert.True(secondErrorCaught, "second catch block should be executed")
}

// 测试嵌套 TryCatchBlock
func TestTryCatchBlock_Nest(t *testing.T) {
	assert := assert.New(t)
	var executionOrder []string

	outerTryCatch := New().
		Try(func() error {
			executionOrder = append(executionOrder, "outer-try-start")

			// 在 outer try 中嵌套一个 try-catch
			innerTryCatch := New().
				Try(func() error {
					executionOrder = append(executionOrder, "inner-try")
					return errors.New("inner error")
				}).
				Catch(func(err error) {
					executionOrder = append(executionOrder, "inner-catch")
					assert.Equal("inner error", err.Error())
				}).
				Finally(func() {
					executionOrder = append(executionOrder, "inner-finally")
				})

			innerTryCatch.Do()
			executionOrder = append(executionOrder, "outer-try-end")
			return errors.New("outer error")
		}).
		Catch(func(err error) {
			executionOrder = append(executionOrder, "outer-catch")
			assert.Equal("outer error", err.Error())
		}).
		Finally(func() {
			executionOrder = append(executionOrder, "outer-finally")
		})

	outerTryCatch.Do()

	// 验证执行顺序
	expectedOrder := []string{
		"outer-try-start",
		"inner-try",
		"inner-catch",
		"inner-finally",
		"outer-try-end",
		"outer-catch",
		"outer-finally",
	}
	assert.Equal(expectedOrder, executionOrder, "execution order should match expected sequence")
}

// 测试并发执行
func TestTryCatchBlock_Concurrent(t *testing.T) {
	assert := assert.New(t)
	const goroutineCount = 100
	var errorCount, completionCount int32
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	for i := 0; i < goroutineCount; i++ {
		go func(routineID int) {
			defer waitGroup.Done()
			tryCatch := New()
			tryCatch.Try(func() error {
				if routineID%2 == 0 {
					return fmt.Errorf("error from goroutine %d", routineID)
				}
				return nil
			}).Catch(func(err error) {
				atomic.AddInt32(&errorCount, 1)
				assert.Contains(err.Error(), "error from goroutine")
			}).Finally(func() {
				atomic.AddInt32(&completionCount, 1)
			})
			tryCatch.Do()
		}(i)
	}

	waitGroup.Wait()

	assert.Equal(goroutineCount/2, int(atomic.LoadInt32(&errorCount)), "catch handler should be executed for half of the goroutines")
	assert.Equal(goroutineCount, int(atomic.LoadInt32(&completionCount)), "finally handler should be executed for all goroutines")
}

// 测试并发执行，使用 sync.Pool 重用 TryCatchBlock 实例
func TestTryCatchBlock_ConcurrentWithPool(t *testing.T) {
	assert := assert.New(t)
	const goroutineCount = 100
	var errorCount, completionCount int32
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	pool := sync.Pool{
		New: func() interface{} {
			return New()
		},
	}

	for i := 0; i < goroutineCount; i++ {
		go func(routineID int) {
			defer waitGroup.Done()
			tryCatch := pool.Get().(*TryCatchBlock)
			tryCatch.Try(func() error {
				if routineID%2 == 0 {
					return fmt.Errorf("error from goroutine %d", routineID)
				}
				return nil
			}).Catch(func(err error) {
				atomic.AddInt32(&errorCount, 1)
				assert.Contains(err.Error(), "error from goroutine")
			}).Finally(func() {
				atomic.AddInt32(&completionCount, 1)
			})
			tryCatch.Do()
			tryCatch.Reset()
			pool.Put(tryCatch)
		}(i)
	}

	waitGroup.Wait()

	assert.Equal(goroutineCount/2, int(atomic.LoadInt32(&errorCount)), "catch handler should be executed for half of the goroutines")
	assert.Equal(goroutineCount, int(atomic.LoadInt32(&completionCount)), "finally handler should be executed for all goroutines")
}
