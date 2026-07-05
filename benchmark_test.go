package gotrycatch

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
)

// --- Do() benchmarks ---

func BenchmarkDo_NoError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := New().Try(func() error {
			return nil
		}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_Error(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		err := New().Try(func() error {
			return testErr
		}).Catch(func(e error) {}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_ErrorWithFinally(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		err := New().Try(func() error {
			return testErr
		}).Catch(func(e error) {}).Finally(func() {}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_Panic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := New().Try(func() error {
			panic("bench panic")
		}).Catch(func(e error) {}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_Full(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	hooks := Hooks{
		OnTryStart: func() {},
		OnTryEnd:   func(error) {},
		OnCatch:    func(error) {},
		OnFinally:  func() {},
	}
	for i := 0; i < b.N; i++ {
		err := NewWithOptions(WithHooks(hooks)).Try(func() error {
			return testErr
		}).Catch(func(e error) {}).Finally(func() {}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_ContextCancelled(b *testing.B) {
	b.ReportAllocs()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for i := 0; i < b.N; i++ {
		err := NewWithOptions(WithContext(ctx)).Try(func() error {
			return nil
		}).Do()
		runtime.KeepAlive(err)
	}
}

func BenchmarkDo_NilTry(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		err := New().Do()
		runtime.KeepAlive(err)
	}
}

// --- TryWithResult benchmarks ---

func BenchmarkTryWithResult_NoError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryWithResult(func() (int, error) {
			return 42, nil
		})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryWithResult_Error(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		v, err := TryWithResult(func() (int, error) {
			return 0, testErr
		})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryWithResult_Panic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryWithResult(func() (int, error) {
			panic("bench panic")
		})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

// --- TryWithResultAndFinally benchmarks ---

func BenchmarkTryWithResultAndFinally_NoError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryWithResultAndFinally(func() (int, error) {
			return 42, nil
		}, func() {})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryWithResultAndFinally_Panic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryWithResultAndFinally(func() (int, error) {
			panic("bench panic")
		}, func() {})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

// --- TryCatchR benchmarks ---

func BenchmarkTryCatchR_NoError(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryCatchR(func() (int, error) {
			return 42, nil
		}, nil, nil)
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryCatchR_Error(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		v, err := TryCatchR(func() (int, error) {
			return 0, testErr
		}, func(e error) {}, nil)
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryCatchR_Panic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		v, err := TryCatchR(func() (int, error) {
			panic("bench panic")
		}, func(e error) {}, nil)
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

func BenchmarkTryCatchR_Full(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		v, err := TryCatchR(func() (int, error) {
			return 0, testErr
		}, func(e error) {}, func() {})
		runtime.KeepAlive(v)
		runtime.KeepAlive(err)
	}
}

// --- Allocation / Pool benchmarks ---

func BenchmarkNew(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		tc := New()
		runtime.KeepAlive(tc)
	}
}

func BenchmarkPoolReuse(b *testing.B) {
	b.ReportAllocs()
	pool := &sync.Pool{
		New: func() any { return New() },
	}
	for i := 0; i < b.N; i++ {
		tc := pool.Get().(*TryCatchBlock)
		tc.Try(func() error { return nil })
		tc.Do()
		tc.Reset()
		pool.Put(tc)
	}
}

func BenchmarkAllocations_Do_NoError(b *testing.B) {
	b.ReportAllocs()
	tc := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Try(func() error { return nil }).Do()
	}
}

func BenchmarkAllocations_Do_Error(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	tc := New()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc.Try(func() error { return testErr }).Catch(func(e error) {}).Do()
	}
}

// --- Native baseline ---

func BenchmarkNative_DeferRecover(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			return nil
		}()
	}
}

func BenchmarkNative_DeferRecover_WithError(b *testing.B) {
	b.ReportAllocs()
	testErr := errors.New("bench error")
	for i := 0; i < b.N; i++ {
		func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			return testErr
		}()
	}
}

func BenchmarkNative_DeferRecover_Panic(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					if e, ok := r.(error); ok {
						err = e
					}
				}
			}()
			panic("bench panic")
		}()
	}
}

// --- pprof-friendly: high-frequency hot path ---

func BenchmarkHotpath_Loop(b *testing.B) {
	b.ReportAllocs()
	pool := &sync.Pool{
		New: func() any { return New() },
	}
	hooks := Hooks{
		OnTryStart: func() {},
		OnTryEnd:   func(error) {},
	}
	testErr := errors.New("bench error")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tc := pool.Get().(*TryCatchBlock)
		tc.ApplyOptions(WithHooks(hooks))
		tc.Try(func() error {
			if i%2 == 0 {
				return nil
			}
			return testErr
		}).Catch(func(e error) {}).Finally(func() {}).Do()
		tc.Reset()
		pool.Put(tc)
	}
}
