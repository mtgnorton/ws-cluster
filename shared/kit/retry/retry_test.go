package retry

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRetryN(t *testing.T) {
	var err error
	t.Run("first_success", func(t *testing.T) {
		r, execTimes := RetryN(3, func(ctx context.Context) (int, error) {
			return 1, nil
		}, WithErrHandle[int](func(subErr error) (stop bool) {
			err = subErr
			return false
		}))
		if err != nil {
			t.Error(err)
		}
		if r != 1 {
			t.Errorf("expected 1, got %d", r)
		}
		if execTimes != 1 {
			t.Errorf("expected 1, got %d", execTimes)
		}
	})

	t.Run("always_errors", func(t *testing.T) {
		beginTime := time.Now()
		r, execTimes := RetryN(3, func(ctx context.Context) (int, error) {
			return 0, errors.New("always_error")
		}, WithErrHandle[int](func(subErr error) (stop bool) {
			err = subErr
			return false
		}))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if r != 0 {
			t.Errorf("expected 0, got %d", r)
		}
		if execTimes != 3 {
			t.Errorf("expected 3, got %d", execTimes)
		}
		if time.Since(beginTime) <= 700*time.Millisecond {
			t.Errorf("expected time to be greater than 700ms, got %v", time.Since(beginTime))
		}
	})
	t.Run("first_error_second_success", func(t *testing.T) {
		beginTime := time.Now()
		i := 0
		r, execTimes := RetryN(3, func(ctx context.Context) (int, error) {
			i++
			if i == 1 {
				return 0, errors.New("first_error")
			}
			return 1, nil
		}, WithErrHandle[int](func(subErr error) (stop bool) {
			err = subErr
			return false
		}))
		if r != 1 {
			t.Errorf("expected 1, got %d", r)
		}
		if execTimes != 2 {
			t.Errorf("expected 2, got %d", execTimes)
		}
		if time.Since(beginTime) <= 100*time.Millisecond {
			t.Errorf("expected time to be greater than 100ms, got %v", time.Since(beginTime))
		}
	})

	t.Run("first_error_stop", func(t *testing.T) {
		r, execTimes := RetryN(3, func(ctx context.Context) (int, error) {
			return 0, errors.New("first_error")
		}, WithErrHandle[int](func(subErr error) (stop bool) {
			return true
		}))
		if r != 0 {
			t.Errorf("expected 0, got %d", r)
		}
		if execTimes != 1 {
			t.Errorf("expected 1, got %d", execTimes)
		}
	})

	t.Run("context_timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer cancel()
		var err error
		r, _ := RetryN(3, func(ctx context.Context) (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 0, errors.New("first_error")
		}, WithContext[int](ctx), WithErrHandle[int](func(subErr error) (stop bool) {
			err = subErr
			return false
		}))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("expected deadline exceeded, got %v", err)
		}
		if r != 0 {
			t.Errorf("expected 0, got %d", r)
		}
	})

	t.Run("custom_backoff", func(t *testing.T) {
		backoff := NewBackoff(WithMin(10*time.Millisecond), WithMax(100*time.Second), WithFactor(2), WithJitter(true))
		var i int
		var err error
		r, execTimes := RetryN(10, func(ctx context.Context) (int, error) {
			i++
			if i == 8 {
				return 1, nil
			}
			return 0, errors.New("error")
		}, WithErrHandle[int](func(subErr error) (stop bool) {
			err = subErr
			return false
		}), WithBackoff[int](backoff))
		if err == nil {
			t.Error("expected error, got nil")
		}
		if r != 1 {
			t.Errorf("expected 1, got %d", r)
		}
		if execTimes != 8 {
			t.Errorf("expected 8, got %d", execTimes)
		}
	})
}
