package gotrycatch

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryCatchBlock_Do(t *testing.T) {
	var (
		finallyCalled       bool
		finallyCalledNested bool
		finallyCalledPanic  bool
	)

	tests := []struct {
		name           string
		tryFunction    func() error
		catchHandler   func(error)
		finallyHandler func()
		shouldPanic    bool
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
			name:        "Finally function",
			tryFunction: func() error { return nil },
			catchHandler: nil,
			finallyHandler: func() {
				finallyCalled = true
			},
		},
		{
			name:           "Try function is nil",
			tryFunction:    nil,
			catchHandler:   nil,
			finallyHandler: nil,
		},
		{
			name: "Nested panic in catch",
			tryFunction: func() error {
				panic("original panic")
			},
			catchHandler: func(err error) {
				assert.Equal(t, "original panic", err.Error())
				assert.NotNil(t, err)
				panic("panic in catch")
			},
			finallyHandler: func() { finallyCalledNested = true },
			shouldPanic:    true,
		},
		{
			name: "Finally executes after panic",
			tryFunction: func() error {
				panic("panic error")
			},
			catchHandler:   nil,
			finallyHandler: func() { finallyCalledPanic = true },
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
			name: "Finally executes after user defer",
			tryFunction: func() error {
				defer func() {
					// 用户自己的 defer，应该在 finally 之前执行
				}()
				return nil
			},
			catchHandler: nil,
			finallyHandler: func() {
				// finally 必须在用户 defer 之后执行
			},
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
			finallyCalled = false
			finallyCalledNested = false
			finallyCalledPanic = false

			tryCatch := New().
				Try(testCase.tryFunction).
				Catch(testCase.catchHandler).
				Finally(testCase.finallyHandler)

			if testCase.shouldPanic {
				assert.Panics(t, func() { tryCatch.Do() })
				if testCase.name == "Nested panic in catch" {
					assert.True(t, finallyCalledNested, "finally should run even when catch panics")
				}
			} else {
				tryCatch.Do()
				switch testCase.name {
				case "Finally function":
					assert.True(t, finallyCalled, "finally handler should be executed")
				case "Finally executes after panic":
					assert.True(t, finallyCalledPanic, "finally handler should be executed after panic")
				}
			}
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
				// 只使用 atomic 计数器，不在 goroutine 内调用 assert（避免 t.FailNow 导致 goroutine panic）
				atomic.AddInt32(&errorCount, 1)
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
				// 只使用 atomic 计数器，不在 goroutine 内调用 assert（避免 t.FailNow 导致 goroutine panic）
				atomic.AddInt32(&errorCount, 1)
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

func TestTryCatchBlock_FinallyPanic(t *testing.T) {
	t.Run("Finally panic propagates", func(t *testing.T) {
		tryCatch := New().
			Try(func() error {
				return nil
			}).
			Finally(func() {
				panic("panic in finally")
			})

		assert.Panics(t, func() {
			tryCatch.Do()
		})
	})

	t.Run("Finally panic with error in try", func(t *testing.T) {
		tryCatch := New().
			Try(func() error {
				return errors.New("error in try")
			}).
			Catch(func(err error) {
			}).
			Finally(func() {
				panic("panic in finally")
			})

		assert.Panics(t, func() {
			tryCatch.Do()
		})
	})

	t.Run("Finally panic after catch panic", func(t *testing.T) {
		tryCatch := New().
			Try(func() error {
				return errors.New("error in try")
			}).
			Catch(func(err error) {
				panic("panic in catch")
			}).
			Finally(func() {
				panic("panic in finally")
			})

		assert.Panics(t, func() {
			tryCatch.Do()
		})
	})
}

func TestTryCatchBlock_Do_PanicReturnsError(t *testing.T) {
	catchCalled := false
	err := New().
		Try(func() error {
			panic("panic error")
		}).
		Catch(func(err error) {
			catchCalled = true
		}).
		Do()

	assert.Error(t, err, "Do() should return error converted from panic")
	assert.Equal(t, "panic error", err.Error())
	assert.True(t, catchCalled, "catch should be called when try panics")
}

func TestTryCatchBlock_Do_PanicWithErrorType(t *testing.T) {
	myErr := errors.New("typed error")
	var caughtErr error
	err := New().
		Try(func() error {
			panic(myErr)
		}).
		Catch(func(err error) {
			caughtErr = err
		}).
		Do()

	assert.Error(t, err)
	assert.Equal(t, myErr, err, "panic with error type should preserve original error")
	assert.Equal(t, myErr, caughtErr)
}

func TestTryCatchBlock_Reset_ClearsAllFields(t *testing.T) {
	tc := New()
	tc.name = "my-block"
	tc.hooks = Hooks{OnTryStart: func() {}, OnTryEnd: func(error) {}, OnCatch: func(error){}, OnFinally: func(){}}
	tc.ctx = context.Background()
	tc.Try(func() error { return nil }).
		Catch(func(error) {}).
		Finally(func() {})

	tc.Reset()

	assert.Nil(t, tc.ctx, "ctx should be nil after Reset")
	assert.Equal(t, Hooks{}, tc.hooks, "hooks should be zero value after Reset")
	assert.Equal(t, "", tc.name, "name should be empty after Reset")
	assert.Nil(t, tc.try, "try should be nil after Reset")
	assert.Nil(t, tc.catch, "catch should be nil after Reset")
	assert.Nil(t, tc.finally, "finally should be nil after Reset")
}

func TestTryCatchBlock_Do_ContextCancelled_FinallyExecutes(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	onFinallyCalled := false
	finallyCalled := false

	err := New().
		ApplyOptions(
			WithContext(ctx),
			WithHooks(Hooks{
				OnFinally: func() { onFinallyCalled = true },
			}),
		).
		Try(func() error {
			return errors.New("should not execute")
		}).
		Catch(func(err error) {
			t.Error("catch should not be called when context is cancelled")
		}).
		Finally(func() {
			finallyCalled = true
		}).
		Do()

	assert.Equal(t, context.Canceled, err, "should return context.Canceled")
	assert.True(t, onFinallyCalled, "OnFinally should be called even when context is cancelled")
	assert.True(t, finallyCalled, "finally should be called even when context is cancelled")
}

func TestTryCatchBlock_Do_PanicCatchCalled(t *testing.T) {
	var caughtErrInCatch error
	err := New().
		Try(func() error {
			panic("try panic")
		}).
		Catch(func(err error) {
			caughtErrInCatch = err
		}).
		Do()

	assert.Error(t, err)
	assert.Equal(t, "try panic", err.Error())
	assert.NotNil(t, caughtErrInCatch, "catch should receive the panic error")
	assert.Equal(t, "try panic", caughtErrInCatch.Error())
}

func TestTryCatchBlock_Do_NilTryFinallyExecutes(t *testing.T) {
	finallyCalled := false

	err := New().
		Finally(func() {
			finallyCalled = true
		}).
		Do()

	assert.NoError(t, err)
	assert.True(t, finallyCalled, "finally must execute when try is nil")
}

func TestTryCatchBlock_Do_NilTryWithOnFinally(t *testing.T) {
	onFinallyCalled := false

	err := New().
		ApplyOptions(WithHooks(Hooks{
			OnFinally: func() { onFinallyCalled = true },
		})).
		Do()

	assert.NoError(t, err)
	assert.True(t, onFinallyCalled, "OnFinally hook must execute when try is nil")
}

func TestTryCatchBlock_Do_NilTryReturnsNil(t *testing.T) {
	err := New().
		Catch(func(err error) {
			t.Error("catch should not be called when try is nil")
		}).
		Do()

	assert.NoError(t, err, "Do() must return nil when try is nil")
}
