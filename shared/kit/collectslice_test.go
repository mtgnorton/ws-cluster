package kit

import (
	"fmt"
	"testing"
)

func TestCollectionS_Reduce(t *testing.T) {
	sum := CollectS([]int{1, 2, 3}).Reduce(func(carry int, item int, index int) int {
		return carry + item
	}, 0)
	if sum != 6 {
		t.Errorf("expected 6, got %d", sum)
	}

	sum = CollectS([]int{1, 2, 3}).Reduce(func(carry int, item int, index int) int {
		return carry + item
	}, 10)
	if sum != 16 {
		t.Errorf("expected 16, got %d", sum)
	}
}

func TestCollectionS_Take(t *testing.T) {
	r := CollectS([]int{1, 2, 3}).Take(2)
	if !equal(r.All(), []int{1, 2}) {
		t.Errorf("expected [1, 2], got %v", r.All())
	}

	r = CollectS([]int{1, 2, 3}).Take(-1)
	if !equal(r.All(), []int{3}) {
		t.Errorf("expected [3], got %v", r.All())
	}
}

func TestCollectionS_Sort(t *testing.T) {
	r := CollectS([]int{3, 1, 2}).Sort(func(a, b int) bool {
		return a < b
	})
	if !equal(r.All(), []int{1, 2, 3}) {
		t.Errorf("expected [1, 2, 3], got %v", r.All())
	}

	r = CollectS([]int{3, 1, 2}).Sort(func(a, b int) bool {
		return a > b
	})
	if !equal(r.All(), []int{3, 2, 1}) {
		t.Errorf("expected [3, 2, 1], got %v", r.All())
	}
}

func TestCollectionS_As(t *testing.T) {

	c := MustAs[string](CollectS([]int{1, 2, 3}).MapAny(func(item, index int) any {
		return fmt.Sprintf("%d", item)
	}))
	fmt.Println(c.Json())
	if !equal(c.All(), []string{"1", "2", "3"}) {
		t.Errorf("expected [1, 2, 3], got %v", c.All())
	}
}

func Test_MapSConc(t *testing.T) {
	r := MapSConc(CollectS([]int{1, 2, 3}), func(item int, index int) string {
		return fmt.Sprintf("%d", item)
	}, 2)
	fmt.Println(r.Json())
}

func TestCollectionS_Unique(t *testing.T) {
	r := CollectS([]int{1, 2, 2, 3, 3, 3}).Unique(func(item int, index int) any {
		return item
	})
	if !equal(r.All(), []int{1, 2, 3}) {
		t.Errorf("expected [1, 2, 3], got %v", r.All())
	}
}

func equal[T comparable](a, b []T) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
