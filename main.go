package main

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

const (
	M = 2
	N = 8
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*50)
	defer cancel()

	result := make([]int, N+1)
	errCh := make(chan error, 1)
	var firstSendErr int32
	wg := new(sync.WaitGroup)
	done := make(chan struct{}, 1)
	limit := make(chan struct{}, M)

	for i := 1; i <= N; i++ {
		limit <- struct{}{}
		var quit bool
		select {
		case <-ctx.Done(): // context已经被cancel，不需要起新的goroutine了
			quit = true
		default:
		}
		if quit {
			break
		}
		wg.Add(1)
		go func(x int) {
			defer func() {
				wg.Done()
				<-limit
			}()
			if ret, err := doTask(ctx, x); err != nil {
				if atomic.CompareAndSwapInt32(&firstSendErr, 0, 1) {
					errCh <- err
					cancel() // cancel其他的请求
				}
			} else {
				result[x] = ret
			}
		}(i)
	}

	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		handleErr(err, result[1:])
		<-done
	case <-done:
		if len(errCh) > 0 {
			err := <-errCh
			handleErr(err, result[1:])
			return
		}
		fmt.Println("success handle all task:", result[1:])
	}
}

func handleErr(err error, result []int) {
	fmt.Println("task err occurs: ", err, "result", result)
}

func doTask(ctx context.Context, i int) (ret int, err error) {
	fmt.Println("task start", i)
	defer func() {
		fmt.Println("task done", i, "err", err)
	}()
	select {
	case <-time.After(time.Second * time.Duration(i)): // 模拟处理任务时间
	case <-ctx.Done(): // 处理任务要支持被context cancel，不然就一直等到处理完再返回了。
		fmt.Println("task canceled", i)
		return -1, ctx.Err()
	}
	if i == 6 { // 模拟出现错误
		return -1, errors.New("err test")
	}

	return i, nil
}
