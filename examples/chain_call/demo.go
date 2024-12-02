package main

import (
	"fmt"

	gtc "github.com/shengyanli1982/go-trycatch"
)

func main() {
	gtc.New().
		Try(func() error {
			// Your code that might return error or panic
			return fmt.Errorf("something went wrong")
		}).
		Catch(func(err error) {
			fmt.Printf("Caught error: %v\n", err)
		}).
		Finally(func() {
			fmt.Println("Cleanup code here")
		}).
		Do()
}
