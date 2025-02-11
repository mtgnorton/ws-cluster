package buf

import (
	"testing"

	"golang.org/x/exp/rand"
)

// ➜  buf go test -bench=.
// goos: darwin
// goarch: arm64
// pkg: github.com/mtgnorton/kit/buf
// cpu: Apple M1 Pro
// BenchmarkAlloc-8        31156742                42.48 ns/op           24 B/op          1 allocs/op
// BenchmarkTradition-8      452659              2600 ns/op           35265 B/op          1 allocs/op
// PASS
// ok      github.com/mtgnorton/kit/buf    4.511s
func BenchmarkAlloc(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	alloc := NewAllocator()
	for i := 0; i < b.N; i++ {
		bytes := alloc.Get(rand.Intn(65536))
		alloc.Put(bytes)
	}
}

func BenchmarkTradition(b *testing.B) {
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		bytes := make([]byte, rand.Intn(65536))
		_ = bytes
		bytes = nil
	}
}
