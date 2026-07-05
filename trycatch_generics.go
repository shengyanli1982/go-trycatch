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

// TryCatchR 执行带泛型返回值的 try-catch-finally 流程，捕获 panic 并转换为错误
func TryCatchR[T any](fn func() (T, error), catch func(error), finally func()) (result T, err error) {
	var catchPanicErr any

	defer func() {
		if r := recover(); r != nil {
			switch v := r.(type) {
			case error:
				err = v
			default:
				err = fmt.Errorf("%v", v)
			}
		}
		if err != nil && catch != nil {
			func() {
				defer func() { catchPanicErr = recover() }()
				catch(err)
			}()
		}
		if finally != nil {
			finally()
		}
		if catchPanicErr != nil {
			panic(catchPanicErr)
		}
	}()

	return fn()
}
