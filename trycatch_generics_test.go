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
