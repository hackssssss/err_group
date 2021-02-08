package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const base = 5

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*100)
	defer cancel()

	errCh := make(chan error, 1)
	var firstSendErr int32
	wg := new(sync.WaitGroup)
	done := make(chan struct{}, 1)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			if err := doTask(ctx, x+base); err != nil {
				if atomic.CompareAndSwapInt32(&firstSendErr, 0, 1) {
					errCh <- err
					cancel() // cancel其他的请求
				}
			}
		}(i)
	}

	go func() {
		wg.Wait()
		done <- struct{}{}
	}()

	select {
	case err := <-errCh:
		handleErr(err)
	case <-done:
		if len(errCh) > 0 {
			err := <-errCh
			handleErr(err)
			return
		}
		fmt.Println("success handle it")
	}

}

func handleErr(err error) {
	fmt.Println("err occurs", err)
}

func doTask(ctx context.Context, i int) (err error) {
	fmt.Println("task start", i-base)
	select {
	case <-time.After(time.Second * time.Duration(i)):
	case <-ctx.Done():
		fmt.Println("task canceled", i-base)
		return ctx.Err()
	}
	fmt.Println("task done", i-base)
	return nil
}
