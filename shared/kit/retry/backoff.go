// 修改自 https://github.com/jpillora/backoff/blob/v1.0.0/backoff.go
package retry

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

const maxInt64 = float64(math.MaxInt64 - 512)

// Backoff 是一个 time.Duration 计数器，从 Min 开始。每次调用
// Duration 方法后，当前时间会乘以 Factor，但不会超过 Max。
//
// Backoff 通常不是并发安全的，但 ForAttempt 方法可以并发使用。
// 使用默认值, 最小100ms, 最大10s, 每次递增2倍,Jitter为false,
// 10次序列为100ms 200ms 400ms 800ms 1.6s 3.2s 6.4s 10s 10s 10s
type Backoff struct {
	attempt atomic.Uint64 // 尝试次数

	factor float64 // Factor 每次递增步骤的乘数

	jitter bool // jitter 通过随机化退避步骤来缓解竞争

	// min 和 max 是计数器的最小值和最大值
	min, max time.Duration
}

func NewBackoff(opts ...BackoffOption) *Backoff {
	b := &Backoff{
		factor: 2,
		jitter: false,
		min:    100 * time.Millisecond,
		max:    10 * time.Second,
	}
	for _, opt := range opts {
		opt(b)
	}
	return b
}

type BackoffOption func(b *Backoff)

func WithFactor(factor float64) BackoffOption {
	return func(b *Backoff) {
		b.factor = factor
	}
}

func WithJitter(jitter bool) BackoffOption {
	return func(b *Backoff) {
		b.jitter = jitter
	}
}

func WithMin(min time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.min = min
	}
}

func WithMax(max time.Duration) BackoffOption {
	return func(b *Backoff) {
		b.max = max
	}
}

// Duration 返回当前尝试的持续时间，然后递增attempt次数。参见 ForAttempt。
func (b *Backoff) Duration() time.Duration {
	// 这里加1后又减1是为了实现原子递增操作，同时保持尝试次数的正确性。
	// Add(1)原子地将attempt加1并返回新值，然后减1得到当前的尝试次数。
	// 这样可以确保每次调用Duration()时，尝试次数都会递增，并且ForAttempt()使用的是正确的当前尝试次数。
	d := b.ForAttempt(float64(b.attempt.Add(1) - 1))
	return d
}

// ForAttempt 返回特定尝试的持续时间。
// 第一次尝试应该是 0。
//
// ForAttempt 是并发安全的。
func (b *Backoff) ForAttempt(attempt float64) time.Duration {

	min := b.min
	if min <= 0 {
		min = 100 * time.Millisecond
	}
	max := b.max
	if max <= 0 {
		max = 10 * time.Second
	}
	if min >= max {
		return max
	}
	factor := b.factor
	if factor <= 0 {
		factor = 2
	}
	minTime := float64(min)
	duration := minTime * math.Pow(factor, attempt)
	if b.jitter {
		// minTime = 100ms
		// 第一次 duration = 100ms * 2^0 = 100ms, 随机化后 duration = (0-1) * (100-100) + 100, 范围为[100ms-100ms]
		// 第二次 duration = 100ms * 2^1 = 200ms, 随机化后 duration = (0-1) * (200-100) + 100, 范围为[100ms-200ms]
		// 第三次 duration = 100ms * 2^2 = 400ms, 随机化后 duration = (0-1) * (400-100) + 100, 范围为[100ms-400ms]
		duration = rand.Float64()*(duration-minTime) + minTime
	}
	// 确保 float64 不会溢出 int64
	if duration > maxInt64 {
		return max
	}
	dur := time.Duration(duration)

	if dur < min {
		return min
	}
	if dur > max {
		return max
	}
	return dur
}

// Reset 将当前尝试计数器重置为零。
func (b *Backoff) Reset() {
	b.attempt.Store(0)
}

// Attempt 返回当前尝试计数器的值。
func (b *Backoff) Attempt() float64 {
	return float64(b.attempt.Load())
}

// Copy 返回一个具有与原始对象相同约束的 backoff
func (b *Backoff) Copy() *Backoff {
	return &Backoff{
		factor: b.factor,
		jitter: b.jitter,
		min:    b.min,
		max:    b.max,
	}
}
