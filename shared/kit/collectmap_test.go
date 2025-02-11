package kit

import (
	"testing"
)

func TestCollectionM_Chunk(t *testing.T) {
	chunk := CollectM(map[string]int{"a": 1, "b": 2, "c": 3, "d": 4, "e": 5}).Chunk(2)
	if len(chunk) != 3 {
		t.Errorf("expected 3 chunks, got %d", len(chunk))
	}

	chunk = CollectM(map[string]int{"a": 1}).Chunk(2)
	if len(chunk) != 1 {
		t.Errorf("expected 1 chunk, got %d", len(chunk))
	}

}
