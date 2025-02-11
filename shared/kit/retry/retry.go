package retry

import (
	"context"
	"fmt"
	"time"
)

type retry[T any] struct {
	Ctx       context.Context         // 当Ctx设置了超时时间, 则当Ctx超时后, 会停止重试
	ErrHandle func(error) (stop bool) // 错误处理函数,可为空,如果返回true, 则停止重试
	Backoff   *Backoff                // 退避策略,可为空
}

type RetryOption[T any] func(r *retry[T])

func WithContext[T any](ctx context.Context) RetryOption[T] {
	return func(r *retry[T]) {
		r.Ctx = ctx
	}
}

func WithErrHandle[T any](errHandle func(error) (stop bool)) RetryOption[T] {
	return func(r *retry[T]) {
		r.ErrHandle = errHandle
	}
}

func WithBackoff[T any](backoff *Backoff) RetryOption[T] {
	return func(r *retry[T]) {
		r.Backoff = backoff
	}
}

// RetryN 重试n次, 如果n<=0, 则n=1
func RetryN[T any](n int, execFunc func(context.Context) (T, error), opts ...RetryOption[T]) (t T, times int) {
	var err error
	if n <= 0 {
		n = 1
	}
	retry := &retry[T]{
		Ctx:     context.Background(),
		Backoff: NewBackoff(),
	}
	for _, opt := range opts {
		opt(retry)
	}

	for i := 0; i < n; i++ {
		times = i + 1
		t, err = execFunc(retry.Ctx)
		if err == nil {
			return
		}
		if retry.ErrHandle != nil {
			if retry.ErrHandle(err) {
				return
			}
		}
		if waitErr := wait(retry.Ctx, retry.Backoff); waitErr != nil {
			err = waitErr
			if retry.ErrHandle != nil {
				retry.ErrHandle(waitErr)
			}
			return
		}
	}
	return
}

// wait 等待一定时间, 如果ctx超时, 则返回ctx.Err()
func wait(ctx context.Context, backoff *Backoff) (err error) {
	waitTime := backoff.Duration()
	fmt.Println("waitTime", waitTime)
	timer := time.NewTimer(waitTime)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}
