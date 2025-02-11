package retry

import (
	"math"
	"testing"
	"time"
)

func TestBackoff(t *testing.T) {
	t.Run("default_backoff", func(t *testing.T) {
		backoff := NewBackoff()
		// 100ms 200ms 400ms 800ms 1.6s 3.2s 6.4s 10s 10s 10s
		expectedDurations := []time.Duration{
			100 * time.Millisecond,
			200 * time.Millisecond,
			400 * time.Millisecond,
			800 * time.Millisecond,
			1600 * time.Millisecond,
			3200 * time.Millisecond,
			6400 * time.Millisecond,
			10 * time.Second,
			10 * time.Second,
			10 * time.Second,
		}
		for i := 0; i < 10; i++ {
			d := backoff.Duration()
			if d != expectedDurations[i] {
				t.Errorf("expected %v, got %v", expectedDurations[i], d)
			}
		}
	})
	t.Run("custom_backoff_no_jitter", func(t *testing.T) {
		backoff := NewBackoff(WithFactor(3), WithMax(time.Hour*24), WithMin(10*time.Millisecond))
		// expectedDurations := []time.Duration{
		// 	10 * time.Millisecond,
		// 	30 * time.Millisecond,
		// 	90 * time.Millisecond,
		// 	270 * time.Millisecond,
		// 	810 * time.Millisecond,
		// 	2430 * time.Millisecond,
		// 	7290 * time.Millisecond,
		// 	21870 * time.Millisecond,
		// 	65610 * time.Millisecond,
		// 	196830 * time.Millisecond,
		// 	590490 * time.Millisecond,
		// 	1771470 * time.Millisecond,
		// 	5314410 * time.Millisecond,
		// 	15943230 * time.Millisecond,
		// 	47829690 * time.Millisecond,
		// 	143489070 * time.Millisecond,
		// 	430467210 * time.Millisecond,
		// 	1291401630 * time.Millisecond,
		// 	3874204890 * time.Millisecond,
		// 	11622614670 * time.Millisecond,
		// 	34867844010 * time.Millisecond,
		// 	104603532030 * time.Millisecond,

		// }
		// [10ms 30ms 90ms 270ms 810ms 2.43s 7.29s 21.87s 1m5.61s 3m16.83s 9m50.49s 29m31.47s 1h28m34.41s
		// 4h25m43.23s 13h17m9.69s 24h0m0s 24h0m0s 24h0m0s 24h0m0s 24h0m0s]
		durations := []time.Duration{
			10 * time.Millisecond,
			30 * time.Millisecond,
			90 * time.Millisecond,
			270 * time.Millisecond,
			810 * time.Millisecond,
			2430 * time.Millisecond,
			7290 * time.Millisecond,
			21870 * time.Millisecond,
			time.Minute + 5*time.Second + 610*time.Millisecond,
			3*time.Minute + 16*time.Second + 830*time.Millisecond,
			9*time.Minute + 50*time.Second + 490*time.Millisecond,
			29*time.Minute + 31*time.Second + 470*time.Millisecond,
			1*time.Hour + 28*time.Minute + 34*time.Second + 410*time.Millisecond,
			4*time.Hour + 25*time.Minute + 43*time.Second + 230*time.Millisecond,
			13*time.Hour + 17*time.Minute + 9*time.Second + 690*time.Millisecond,
			24 * time.Hour,
			24 * time.Hour,
			24 * time.Hour,
			24 * time.Hour,
			24 * time.Hour,
		}
		for i := 0; i < 20; i++ {
			d := backoff.Duration()
			if d != durations[i] {
				t.Errorf("expected %v, got %v", durations[i], d)
			}
		}
	})

	t.Run("custom_backoff_with_jitter", func(t *testing.T) {
		min := 100 * time.Millisecond
		backoff := NewBackoff(WithJitter(true), WithMin(min))
		// 100ms 200ms 400ms 800ms 1.6s 3.2s 6.4s
		// expectedDurations := []time.Duration{
		// 	100 * time.Millisecond,
		// 	200 * time.Millisecond,
		// 	400 * time.Millisecond,
		// 	800 * time.Millisecond,
		// 	1600 * time.Millisecond,
		// 	3200 * time.Millisecond,
		// 	6400 * time.Millisecond,
		// }
		for i := 0; i < 7; i++ {
			d := backoff.Duration()
			currentMax := float64(min) * math.Pow(2, float64(i))
			if d < min || d > time.Duration(currentMax) {
				t.Errorf("expected %v, got %v", currentMax, d)
			}
		}
	})
}
