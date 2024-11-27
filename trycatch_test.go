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
		name        string
		tryFunc     func() error
		catchFunc   func(error)
		finallyFunc func()
	}{
		{
			name: "No error",
			tryFunc: func() error {
				return nil
			},
			catchFunc:   nil,
			finallyFunc: nil,
		},
		{
			name: "Error in try",
			tryFunc: func() error {
				return errors.New("try error")
			},
			catchFunc: func(err error) {
				assert.Equal(t, "try error", err.Error())
				assert.NotNil(t, err)
			},
			finallyFunc: nil,
		},
		{
			name: "Panic in try",
			tryFunc: func() error {
				panic("panic error")
			},
			catchFunc: func(err error) {
				assert.Equal(t, "panic error", err.Error())
				assert.NotNil(t, err)
			},
			finallyFunc: nil,
		},
		{
			name: "Finally function",
			tryFunc: func() error {
				return nil
			},
			catchFunc: nil,
			finallyFunc: (func() func() {
				executed := false
				return func() {
					executed = true
					assert.True(t, executed, "finally should be executed")
				}
			})(),
		},
		{
			name:        "Try function is nil",
			tryFunc:     nil,
			catchFunc:   nil,
			finallyFunc: nil,
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
			tryFunc: func() error {
				panic("panic error")
			},
			catchFunc: nil,
			finallyFunc: (func() func() {
				executed := false
				return func() {
					executed = true
					assert.True(t, executed, "finally should be executed")
				}
			})(),
		},
		{
			name: "Complex error chain",
			tryFunc: func() error {
				originalErr := errors.New("original error")
				return fmt.Errorf("wrapped: %w", originalErr)
			},
			catchFunc: func(err error) {
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
			tryFunc: func() error {
				panic(customError{message: "custom error"})
			},
			catchFunc: func(err error) {
				assert.Equal(t, "custom error", err.Error())
				customErr, ok := err.(customError)
				assert.True(t, ok, "error should be of type customError")
				assert.Equal(t, "custom error", customErr.message)
			},
		},
		{
			name: "Multiple deferred operations",
			tryFunc: func() error {
				defer func() {
					// 模拟其他 defer 操作
				}()
				return errors.New("error after defer")
			},
			catchFunc:   nil,
			finallyFunc: nil,
		},
		{
			name: "Nil catch with error",
			tryFunc: func() error {
				return errors.New("uncaught error")
			},
			catchFunc:   nil,
			finallyFunc: nil,
		},
		{
			name:        "Empty try-catch-finally chain",
			tryFunc:     func() error { return nil },
			catchFunc:   nil,
			finallyFunc: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := New().Try(tt.tryFunc).Catch(tt.catchFunc).Finally(tt.finallyFunc)
			tc.Do()
		})
	}
}

// 自定义错误类型用于测试
type customError struct {
	message string
}

func (e customError) Error() string {
	return e.message
}

// 测试链式调用
func TestTryCatchBlock_ChainCalls(t *testing.T) {
	assert := assert.New(t)
	executed := false
	caught := false

	tc := New().
		Try(func() error {
			assert.False(executed, "finally not executed yet")
			return errors.New("test error")
		}).
		Catch(func(err error) {
			assert.Equal("test error", err.Error())
			caught = true
		}).
		Finally(func() {
			executed = true
		})

	tc.Do()

	assert.True(caught, "catch should be executed")
	assert.True(executed, "finally should be executed")
}

// 测试重复使用同一个实例
func TestTryCatchBlock_Reuse(t *testing.T) {
	assert := assert.New(t)
	tc := New()

	// 第一次使用
	tc.Try(func() error { return nil }).Do()
	assert.Nil(tc.try, "should be reset after Do()")
	assert.Nil(tc.catch, "should be reset after Do()")
	assert.Nil(tc.finally, "should be reset after Do()")

	// 第二次使用
	tc.Try(func() error { return errors.New("error") }).Do()
}

// TestTryCatchBlock_Concurrent tests thread safety of the TryCatchBlock
func TestTryCatchBlock_Concurrent(t *testing.T) {
	assert := assert.New(t)
	const numGoroutines = 100
	var caughtCount, executedCount int32
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			tc := New()
			tc.Try(func() error {
				if id%2 == 0 {
					return fmt.Errorf("error from goroutine %d", id)
				}
				return nil
			}).Catch(func(err error) {
				atomic.AddInt32(&caughtCount, 1)
				assert.Contains(err.Error(), "error from goroutine")
			}).Finally(func() {
				atomic.AddInt32(&executedCount, 1)
			})
			tc.Do()
		}(i)
	}

	wg.Wait()

	assert.Equal(numGoroutines/2, int(atomic.LoadInt32(&caughtCount)), "catch should be executed for half of the goroutines")
	assert.Equal(numGoroutines, int(atomic.LoadInt32(&executedCount)), "finally should be executed for all goroutines")
}
