package lru

import (
	"fmt"
	"testing"
)

func TestNewARC(t *testing.T) {
	a, err := NewARC[int, int](128)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	for i := 0; i < 256; i++ {
		a.Add(i, i)
	}
	if a.Len() != 128 {
		t.Fatalf("bad: %d", a.Len())
	}
	if a.Cap() != 128 {
		t.Fatalf("bad: %d", a.Cap())
	}

	for i, k := range a.Keys() {
		if v, ok := a.Get(k); !ok || v != k || v != i+128 {
			fmt.Println(k)
			t.Fatalf("bad key: %v", k)
		}
	}
	for i, v := range a.Values() {
		if v != i+128 {
			t.Fatalf("bad value: %v", v)
		}
	}
	for i := 0; i < 128; i++ {
		if _, ok := a.Get(i); ok {
			t.Fatalf("should be evicted")
		}
	}
	for i := 128; i < 256; i++ {
		if _, ok := a.Get(i); !ok {
			t.Fatalf("should not be evicted")
		}
	}
	for i := 128; i < 192; i++ {
		a.Remove(i)
		if _, ok := a.Get(i); ok {
			t.Fatalf("should be deleted")
		}
	}
	if a.Cap() != 128 {
		t.Fatalf("expect %d, but %d", 128, a.Cap())
	}

}

func TestARC_Adaptive(t *testing.T) {
	l, err := NewARC[int, int](4)
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	// Fill t1
	for i := 0; i < 4; i++ {
		l.Add(i, i)
	}
	if n := l.t1.Len(); n != 4 {
		t.Fatalf("bad: %d", n)
	}
	// Move to t2
	l.Get(0)
	l.Get(1)
	if n := l.t2.Len(); n != 2 {
		t.Fatalf("bad: %d", n)
	}
	fmt.Println(l)

	// Evict from t1
	l.Add(4, 4)

	fmt.Println(l)

	if n := l.b1.Len(); n != 1 {
		t.Fatalf("bad: %d", n)
	}

	// Current state
	// t1 : (MRU) [4, 3] (LRU)
	// t2 : (MRU) [1, 0] (LRU)
	// b1 : (MRU) [2] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 2, should cause hit on b1
	l.Add(2, 2)
	if n := l.b1.Len(); n != 1 {
		t.Fatalf("bad: %d", n)
	}
	if l.p != 1 {
		t.Fatalf("bad: %d", l.p)
	}
	if n := l.t2.Len(); n != 3 {
		t.Fatalf("bad: %d", n)
	}

	// Current state
	// t1 : (MRU) [4] (LRU)
	// t2 : (MRU) [2, 1, 0] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 4, should migrate to t2
	l.Add(4, 4)
	if n := l.t1.Len(); n != 0 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.t2.Len(); n != 4 {
		t.Fatalf("bad: %d", n)
	}

	// Current state
	// t1 : (MRU) [] (LRU)
	// t2 : (MRU) [4, 2, 1, 0] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [] (LRU)

	// Add 4, should evict to b2
	l.Add(5, 5)
	if n := l.t1.Len(); n != 1 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.t2.Len(); n != 3 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.b2.Len(); n != 1 {
		t.Fatalf("bad: %d", n)
	}

	// Current state
	// t1 : (MRU) [5] (LRU)
	// t2 : (MRU) [4, 2, 1] (LRU)
	// b1 : (MRU) [3] (LRU)
	// b2 : (MRU) [0] (LRU)

	// Add 0, should decrease p
	l.Add(0, 0)
	if n := l.t1.Len(); n != 0 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.t2.Len(); n != 4 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.b1.Len(); n != 2 {
		t.Fatalf("bad: %d", n)
	}
	if n := l.b2.Len(); n != 0 {
		t.Fatalf("bad: %d", n)
	}
	if l.p != 0 {
		t.Fatalf("bad: %d", l.p)
	}

	// Current state
	// t1 : (MRU) [] (LRU)
	// t2 : (MRU) [0, 4, 2, 1] (LRU)
	// b1 : (MRU) [5, 3] (LRU)
	// b2 : (MRU) [0] (LRU)
}

func TestARC_EvictCallBack(t *testing.T) {

	var callbackValues = []int{2, 3, 0}
	var callbackIndex = 0
	l, err := NewARC[int, int](4, func(key int, value int) {
		if value != callbackValues[callbackIndex] {
			t.Fatalf("bad: %d", value)
		}
		callbackIndex++
	})
	if err != nil {
		t.Fatalf("err: %v", err)
	}

	for i := 0; i < 4; i++ {
		l.Add(i, i)
	}
	// Fill t1 后
	// ARC{size=4,len:4, p=0,
	// t1=Simple LRU: [3:3 2:2 1:1 0:0],
	// t2=Simple LRU: [],
	// b1=Simple LRU: [],
	// b2=Simple LRU: []}

	if n := l.t1.Len(); n != 4 {
		t.Fatalf("bad: %d", n)
	}

	// Move to t2
	l.Get(0)
	l.Get(1)
	// 获取0,1后
	// ARC{size=4,len:4, p=0,
	// t1=Simple LRU: [3:3 2:2],
	// t2=Simple LRU: [1:1 0:0],
	// b1=Simple LRU: [],
	// b2=Simple LRU: []}

	if n := l.t2.Len(); n != 2 {
		t.Fatalf("bad: %d", n)
	}

	// Evict from t1
	l.Add(4, 4)
	// 添加4后
	// ARC{size=4,len:4, p=0,
	// t1=Simple LRU: [4:4 3:3],
	// t2=Simple LRU: [1:1 0:0],
	// b1=Simple LRU: [2:{}],
	// b2=Simple LRU: []}
	if n := l.b1.Len(); n != 1 {
		t.Fatalf("bad: %d", n)
	}

	l.Add(2, 2)
	// 添加2后
	// ARC{size=4,len:4, p=1,
	// t1=Simple LRU: [4:4],
	// t2=Simple LRU: [2:2 1:1 0:0],
	// b1=Simple LRU: [3:{}],
	// b2=Simple LRU: []}

	l.Add(5, 5)

	// 添加5后
	// ARC{size=4,len:4, p=1,
	// t1=Simple LRU: [5:5 4:4],
	// t2=Simple LRU: [2:2 1:1],
	// b1=Simple LRU: [3:{}],
	// b2=Simple LRU: [0:{}]}
}
