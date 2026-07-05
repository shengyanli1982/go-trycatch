package gotrycatch

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWithContext_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := New().
		ApplyOptions(WithContext(ctx)).
		Try(func() error {
			return errors.New("should not execute")
		}).
		Do()

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestWithContext_Timeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Test that context is checked before execution
	start := time.Now()
	err := New().
		ApplyOptions(WithContext(ctx)).
		Try(func() error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}).
		Do()

	elapsed := time.Since(start)
	// With current implementation, context is only checked once at start
	// so the sleep will complete and no error will be returned
	assert.NoError(t, err)
	assert.True(t, elapsed >= 50*time.Millisecond, "should have waited at least timeout duration")
}

func TestWithContext_CancelledBeforeExecution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := New().
		ApplyOptions(WithContext(ctx)).
		Try(func() error {
			return errors.New("should not execute")
		}).
		Do()

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestWithContext_NoCancellation(t *testing.T) {
	ctx := context.Background()

	err := New().
		ApplyOptions(WithContext(ctx)).
		Try(func() error {
			return nil
		}).
		Do()

	assert.NoError(t, err)
}

func TestHooks_OnTryStart(t *testing.T) {
	tryStartCalled := false

	New().
		ApplyOptions(WithHooks(Hooks{
			OnTryStart: func() {
				tryStartCalled = true
			},
		})).
		Try(func() error {
			return nil
		}).
		Do()

	assert.True(t, tryStartCalled, "OnTryStart should be called")
}

func TestHooks_OnTryEnd(t *testing.T) {
	var capturedErr error

	New().
		ApplyOptions(WithHooks(Hooks{
			OnTryEnd: func(err error) {
				capturedErr = err
			},
		})).
		Try(func() error {
			return errors.New("test error")
		}).
		Do()

	assert.Error(t, capturedErr)
	assert.Equal(t, "test error", capturedErr.Error())
}

func TestHooks_OnCatch(t *testing.T) {
	var caughtErr error

	New().
		ApplyOptions(WithHooks(Hooks{
			OnCatch: func(err error) {
				caughtErr = err
			},
		})).
		Try(func() error {
			return errors.New("test error")
		}).
		Catch(func(err error) {
			// User's catch handler
		}).
		Do()

	// OnCatch is called when user's Catch handler is called
	// The error is "test error" passed to the catch handler
	assert.Equal(t, "test error", caughtErr.Error())
}

func TestHooks_OnFinally(t *testing.T) {
	finallyCalled := false

	New().
		ApplyOptions(WithHooks(Hooks{
			OnFinally: func() {
				finallyCalled = true
			},
		})).
		Try(func() error {
			return nil
		}).
		Finally(func() {
			// User's finally handler
		}).
		Do()

	assert.True(t, finallyCalled, "OnFinally should be called")
}

func TestHooks_ExecutionOrder(t *testing.T) {
	var order []string

	New().
		ApplyOptions(WithHooks(Hooks{
			OnTryStart: func() {
				order = append(order, "on-try-start")
			},
			OnTryEnd: func(err error) {
				order = append(order, "on-try-end")
			},
			OnCatch: func(err error) {
				order = append(order, "on-catch")
			},
			OnFinally: func() {
				order = append(order, "on-finally")
			},
		})).
		Try(func() error {
			order = append(order, "try")
			return nil
		}).
		Catch(func(err error) {
			order = append(order, "catch")
		}).
		Finally(func() {
			order = append(order, "finally")
		}).
		Do()

	// Execution order: on-try-start -> try -> on-try-end -> on-finally -> finally
	expectedOrder := []string{
		"on-try-start",
		"try",
		"on-try-end",
		"on-finally",
		"finally",
	}
	assert.Equal(t, expectedOrder, order)
}

func TestHooks_ExecutionOrderWithError(t *testing.T) {
	var order []string

	New().
		ApplyOptions(WithHooks(Hooks{
			OnTryStart: func() {
				order = append(order, "on-try-start")
			},
			OnTryEnd: func(err error) {
				order = append(order, "on-try-end")
			},
			OnCatch: func(err error) {
				order = append(order, "on-catch")
			},
			OnFinally: func() {
				order = append(order, "on-finally")
			},
		})).
		Try(func() error {
			order = append(order, "try")
			return errors.New("test error")
		}).
		Catch(func(err error) {
			order = append(order, "catch")
		}).
		Finally(func() {
			order = append(order, "finally")
		}).
		Do()

	// Execution order: on-try-start -> try -> on-try-end -> on-catch -> catch -> on-finally -> finally
	expectedOrder := []string{
		"on-try-start",
		"try",
		"on-try-end",
		"on-catch",
		"catch",
		"on-finally",
		"finally",
	}
	assert.Equal(t, expectedOrder, order)
}

func TestWithName(t *testing.T) {
	tc := New()
	WithName("my-block")(tc)
	assert.Equal(t, "my-block", tc.Name())
}

func TestContextGetter(t *testing.T) {
	ctx := context.Background()
	tc := New()
	tc.ctx = ctx
	assert.Equal(t, ctx, tc.Context())
}

func TestApplyOptions(t *testing.T) {
	ctx := context.Background()
	tc := New()
	tc.ApplyOptions(
		WithContext(ctx),
		WithName("test-block"),
	)
	assert.Equal(t, ctx, tc.ctx)
	assert.Equal(t, "test-block", tc.name)
}

func TestNewWithOptions_Single(t *testing.T) {
	ctx := context.Background()
	tc := NewWithOptions(WithContext(ctx))
	assert.Equal(t, ctx, tc.ctx)
	assert.NotNil(t, tc)
}

func TestNewWithOptions_Multiple(t *testing.T) {
	ctx := context.Background()
	tc := NewWithOptions(
		WithContext(ctx),
		WithName("multi-block"),
		WithHooks(Hooks{
			OnTryStart: func() {},
		}),
	)

	assert.Equal(t, ctx, tc.ctx)
	assert.Equal(t, "multi-block", tc.Name())
	assert.NotNil(t, tc.Hooks().OnTryStart)
}

func TestNewWithOptions_NoOptions(t *testing.T) {
	tc := NewWithOptions()
	assert.NotNil(t, tc)
	assert.Nil(t, tc.ctx)
	assert.Equal(t, "", tc.Name())
	assert.Equal(t, Hooks{}, tc.Hooks())
}
