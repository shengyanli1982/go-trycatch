package main

import (
	"fmt"
	"sync"

	gtc "github.com/shengyanli1982/go-trycatch"
)

func main() {
	// Number of goroutines to create
	const goroutineCount = 100
	// WaitGroup to synchronize all goroutines
	var waitGroup sync.WaitGroup
	waitGroup.Add(goroutineCount)

	// Create a sync.Pool to reuse TryCatchBlock instances
	// This helps reduce memory allocations in concurrent scenarios
	pool := sync.Pool{
		New: func() interface{} {
			return gtc.New()
		},
	}

	// Launch goroutines
	for i := 0; i < goroutineCount; i++ {
		go func(routineID int) {
			// Ensure WaitGroup is decremented when goroutine completes
			defer waitGroup.Done()

			// Get a TryCatchBlock instance from the pool
			tryCatch := pool.Get().(*gtc.TryCatchBlock)

			// Execute the try-catch-finally block
			tryCatch.Try(func() error {
				// Simulate error for even-numbered routines
				if routineID%2 == 0 {
					return fmt.Errorf("error from goroutine %d", routineID)
				}
				return nil
			}).Catch(func(err error) {
				// Handle any errors that occurred
				fmt.Printf("Caught error from goroutine %d: %v\n", routineID, err)
			}).Finally(func() {
				// Cleanup code that runs regardless of error status
				fmt.Printf("Goroutine %d completed\n", routineID)
			}).Do()

			// Reset the TryCatchBlock before returning it to the pool
			tryCatch.Reset()
			pool.Put(tryCatch)
		}(i)
	}
}
