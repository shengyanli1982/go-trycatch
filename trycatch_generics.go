package gotrycatch

import (
	"fmt"
)

// TryWithResult 执行带返回值的函数，捕获 panic 并转换为错误
func TryWithResult[T any](fn func() (T, error)) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}
	}()

	return fn()
}

// TryWithResultAndFinally 类似 TryWithResult，但额外接受 finally 处理器
func TryWithResultAndFinally[T any](fn func() (T, error), finally func()) (result T, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}
		if finally != nil {
			finally()
		}
	}()

	return fn()
}
