package shared

import (
	"fmt"
	"testing"
	"time"
)

func exec2S(timeout time.Duration) {
	end := TimeoutDetection.Do(timeout, func() {
		fmt.Println("exec2S timeout")
	})
	defer end()
	time.Sleep(time.Second)
	fmt.Println("exec2S")
}
func TestTimeoutDetection_Do(t *testing.T) {
	t.Run("no timeout", func(t *testing.T) {
		exec2S(2 * time.Second)
	})
	t.Run("timeout", func(t *testing.T) {
		exec2S(time.Millisecond * 500)

	})
}
