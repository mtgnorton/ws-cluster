package lru

import (
	"testing"
)

func TestSimple(t *testing.T) {
	lru := testSimpleInstance(t)

	lru.Add(4, 4)

	// 4 3 2
	if lru.ll.Len() != 3 {
		t.Fatalf("expected len 3, got %d", lru.ll.Len())
	}
	if k, v, isFound := lru.GetOldest(); !isFound || k != 2 || v != 2 {
		t.Fatalf("expected key 2, value 2, got key %d, value %d", k, v)
	}
	if k, v, isFound := lru.RemoveOldest(); !isFound || k != 2 || v != 2 {
		t.Fatalf("expected key 2, value 2, got key %d, value %d", k, v)
	}
	if lru.ll.Len() != 2 {
		t.Fatalf("expected len 2, got %d", lru.ll.Len())
	}
}

func TestSimple_Peek(t *testing.T) {

	lru := testSimpleInstance(t)
	if v, ok := lru.Peek(2); !ok || v != 2 {
		t.Fatalf("expected value 2, got %d", v)
	}
	if k, v, ok := lru.GetNewest(); !ok || k != 3 || v != 3 {
		t.Fatalf("expected key 3, value 3, got key %d, value %d", k, v)
	}
	if k, v, ok := lru.GetOldest(); !ok || k != 1 || v != 1 {
		t.Fatalf("expected key 1, value 1, got key %d, value %d", k, v)
	}
}

func TestSimple_Purge(t *testing.T) {
	lru := testSimpleInstance(t)
	lru.Purge()
	if lru.ll.Len() != 0 {
		t.Fatalf("expected len 0, got %d", lru.ll.Len())
	}
	if _, ok := lru.Get(3); ok {
		t.Fatalf("should contain nothing")
	}
}
func TestSimple_Resize(t *testing.T) {
	lru := testSimpleInstance(t)
	evict := lru.Resize(2)
	if evict != 1 {
		t.Fatalf("expected evict 1, got %d", evict)
	}
	if lru.String() != "Simple LRU: [3:3 2:2]" {
		t.Fatalf("expected 'Simple LRU: [3:3 2:2]', got %s", lru.String())
	}
}

func TestSimple_Get(t *testing.T) {
	lru := testSimpleInstance(t)
	if v, ok := lru.Get(2); !ok || v != 2 {
		t.Fatalf("expected value 2, got %d", v)
	}
	if v, ok := lru.Get(4); ok {
		t.Fatalf("expected not found, got %d", v)
	}
}

func testSimpleInstance(t *testing.T) *Simple[int, int] {
	lru, err := NewSimple[int, int](3)
	if err != nil {
		t.Fatal(err)
	}
	if lru == nil {
		t.Fatal("lru is nil")
	}
	if lru.size != 3 {
		t.Fatalf("expected size 3, got %d", lru.size)
	}
	if lru.items == nil {
		t.Fatal("items is nil")
	}
	if lru.ll == nil {
		t.Fatal("ll is nil")
	}
	lru.Add(1, 1)
	lru.Add(2, 2)
	lru.Add(3, 3)
	return lru
}
