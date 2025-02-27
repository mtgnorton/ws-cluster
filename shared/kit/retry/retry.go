package retry

import (
	"context"
	"time"
)

type retry struct {
	Ctx       context.Context         // 当Ctx设置了超时时间, 则当Ctx超时后, 会停止重试
	ErrHandle func(error) (stop bool) // 错误处理函数,可为空,如果返回true, 则停止重试
	Backoff   *Backoff                // 退避策略,可为空
}

type RetryOption func(r *retry)

func WithContext(ctx context.Context) RetryOption {
	return func(r *retry) {
		r.Ctx = ctx
	}
}

func WithErrHandle(errHandle func(error) (stop bool)) RetryOption {
	return func(r *retry) {
		r.ErrHandle = errHandle
	}
}

func WithBackoff(backoff *Backoff) RetryOption {
	return func(r *retry) {
		r.Backoff = backoff
	}
}

// RetryN 重试n次, 如果n<=0, 则n=1
// 如果ctx设置了超时时间, 则当ctx超时后, 会停止重试
// 如果ErrHandle设置了错误处理函数, 则当错误处理函数返回true时, 会停止重试
// Backoff 默认回退时间为 10次序列为100ms 200ms 400ms 800ms 1.6s 3.2s 6.4s 10s 10s 10s
func RetryN(n int, execFunc func(context.Context) (interface{}, error), opts ...RetryOption) (t interface{}, times int) {
	var err error
	if n <= 0 {
		n = 1
	}
	retry := &retry{
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
	timer := time.NewTimer(waitTime)
	defer timer.Stop()
	select {
	case <-timer.C:
	case <-ctx.Done():
		err = ctx.Err()
	}
	return
}
