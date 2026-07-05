package gotrycatch

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTryWithResult_Success(t *testing.T) {
	result, err := TryWithResult(func() (int, error) {
		return 42, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
}

func TestTryWithResult_WithError(t *testing.T) {
	expectedErr := errors.New("test error")
	result, err := TryWithResult(func() (int, error) {
		return 0, expectedErr
	})

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 0, result)
}

func TestTryWithResult_WithPanic(t *testing.T) {
	result, err := TryWithResult(func() (int, error) {
		panic("panic error")
	})

	assert.Error(t, err, "panic should be converted to error, not re-panicked")
	assert.Equal(t, "panic error", err.Error())
	assert.Equal(t, 0, result, "result should be zero value on panic")
}

func TestTryWithResult_WithTypedError(t *testing.T) {
	myErr := errors.New("typed panic")
	result, err := TryWithResult(func() (int, error) {
		panic(myErr)
	})

	assert.Error(t, err)
	assert.Equal(t, myErr, err, "should preserve original error from panic")
	assert.Equal(t, 0, result)
}

func TestTryWithResultAndFinally_Success(t *testing.T) {
	finallyCalled := false
	result, err := TryWithResultAndFinally(
		func() (int, error) {
			return 42, nil
		},
		func() {
			finallyCalled = true
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.True(t, finallyCalled, "finally should be called")
}

func TestTryWithResultAndFinally_WithError(t *testing.T) {
	finallyCalled := false
	expectedErr := errors.New("test error")
	result, err := TryWithResultAndFinally(
		func() (int, error) {
			return 0, expectedErr
		},
		func() {
			finallyCalled = true
		},
	)

	assert.Error(t, err)
	assert.Equal(t, expectedErr, err)
	assert.Equal(t, 0, result)
	assert.True(t, finallyCalled, "finally should be called even with error")
}

func TestTryWithResultAndFinally_WithPanic(t *testing.T) {
	finallyCalled := false
	result, err := TryWithResultAndFinally(
		func() (int, error) {
			panic("panic error")
		},
		func() {
			finallyCalled = true
		},
	)

	assert.Error(t, err, "panic should be converted to error, not re-panicked")
	assert.Equal(t, "panic error", err.Error())
	assert.Equal(t, 0, result, "result should be zero value on panic")
	assert.True(t, finallyCalled, "finally should be called even with panic")
}

func TestTryWithResult_WithString(t *testing.T) {
	result, err := TryWithResult(func() (string, error) {
		return "hello world", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "hello world", result)
}

func TestTryWithResult_WithSlice(t *testing.T) {
	result, err := TryWithResult(func() ([]int, error) {
		return []int{1, 2, 3}, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, []int{1, 2, 3}, result)
}

func TestTryWithResult_WithStruct(t *testing.T) {
	type Person struct {
		Name string
		Age  int
	}

	result, err := TryWithResult(func() (Person, error) {
		return Person{Name: "Alice", Age: 30}, nil
	})

	assert.NoError(t, err)
	assert.Equal(t, Person{Name: "Alice", Age: 30}, result)
}

func TestTryCatchR_Success(t *testing.T) {
	finallyCalled := false

	result, err := TryCatchR(
		func() (int, error) {
			return 42, nil
		},
		nil,
		func() {
			finallyCalled = true
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, 42, result)
	assert.True(t, finallyCalled, "finally should be called on success")
}

func TestTryCatchR_WithError(t *testing.T) {
	var caughtErr error
	finallyCalled := false

	result, err := TryCatchR(
		func() (string, error) {
			return "", errors.New("try error")
		},
		func(err error) {
			caughtErr = err
		},
		func() {
			finallyCalled = true
		},
	)

	assert.Error(t, err)
	assert.Equal(t, "try error", err.Error())
	assert.NotNil(t, caughtErr)
	assert.Equal(t, "try error", caughtErr.Error())
	assert.Equal(t, "", result)
	assert.True(t, finallyCalled, "finally should be called on error")
}

func TestTryCatchR_WithPanic(t *testing.T) {
	var caughtErr error
	finallyCalled := false

	result, err := TryCatchR(
		func() (int, error) {
			panic("panic in try")
		},
		func(err error) {
			caughtErr = err
		},
		func() {
			finallyCalled = true
		},
	)

	assert.Error(t, err)
	assert.Equal(t, "panic in try", err.Error())
	assert.NotNil(t, caughtErr)
	assert.Equal(t, "panic in try", caughtErr.Error())
	assert.Equal(t, 0, result)
	assert.True(t, finallyCalled, "finally should be called on panic")
}

func TestTryCatchR_FinallyAlwaysRuns(t *testing.T) {
	finallyCount := 0

	// finally on success
	TryCatchR[int](
		func() (int, error) { return 1, nil },
		nil,
		func() { finallyCount++ },
	)

	// finally on error
	TryCatchR[int](
		func() (int, error) { return 0, errors.New("err") },
		nil,
		func() { finallyCount++ },
	)

	// finally on panic
	TryCatchR[int](
		func() (int, error) { panic("p") },
		nil,
		func() { finallyCount++ },
	)

	assert.Equal(t, 3, finallyCount, "finally should run in all three scenarios")
}

func TestTryCatchR_NilCatchNilFinally(t *testing.T) {
	result, err := TryCatchR(
		func() (int, error) {
			return 100, errors.New("no catcher")
		},
		nil,
		nil,
	)

	assert.Error(t, err)
	assert.Equal(t, "no catcher", err.Error())
	assert.Equal(t, 100, result)
}

func TestTryCatchR_CatchPanic(t *testing.T) {
	finallyCalled := false

	assert.Panics(t, func() {
		TryCatchR[int](
			func() (int, error) {
				return 0, errors.New("original error")
			},
			func(err error) {
				panic("panic in catch")
			},
			func() {
				finallyCalled = true
			},
		)
	}, "should propagate catch panic")

	assert.True(t, finallyCalled, "finally should run even when catch panics")
}

func TestTryCatchR_CatchPanicAfterFnPanic(t *testing.T) {
	finallyCount := 0

	assert.Panics(t, func() {
		TryCatchR[int](
			func() (int, error) {
				panic("panic in fn")
			},
			func(err error) {
				panic("panic in catch")
			},
			func() {
				finallyCount++
			},
		)
	}, "should propagate catch panic")

	assert.Equal(t, 1, finallyCount, "finally must be called exactly once when both fn and catch panic")
}
