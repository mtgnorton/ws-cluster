package kit

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

func TestWorkerManager(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// 清理测试数据
	ctx := context.Background()
	rdb.Del(ctx, AvailableKey)
	keys, err := rdb.Keys(ctx, WorkerKeyPrefix+"*").Result()
	if err != nil {
		t.Fatalf("Failed to get worker keys: %v", err)
	}
	for _, key := range keys {
		rdb.Del(ctx, key)
	}

	t.Run("New", func(t *testing.T) {
		w, err := NewNodeIDWorker(rdb)
		if err != nil {
			t.Fatalf("New failed: %v", err)
		}
		if w == nil {
			t.Fatal("New returned nil")
		}
	})

	t.Run("Acquire", func(t *testing.T) {
		w, _ := NewNodeIDWorker(rdb)
		id, err := w.Acquire()
		if err != nil {
			t.Fatalf("Acquire failed: %v", err)
		}
		if id < 0 || id > MaxWorkerID {
			t.Fatalf("Invalid worker ID: %d", id)
		}

		// 测试重复获取
		_, err = w.Acquire()
		if err != ErrAlreadyAcquired {
			t.Fatalf("Expected ErrAlreadyAcquired, got %v", err)
		}

		w.Release()
	})

	t.Run("Release", func(t *testing.T) {
		w, _ := NewNodeIDWorker(rdb)
		id, _ := w.Acquire()

		err := w.Release()
		if err != nil {
			t.Fatalf("Release failed: %v", err)
		}

		// 验证ID已经释放回池中
		exists, err := rdb.SIsMember(ctx, AvailableKey, id).Result()
		if err != nil {
			t.Fatalf("Failed to check if ID is in available set: %v", err)
		}
		if !exists {
			t.Fatalf("ID %d was not released back to the pool", id)
		}
	})

	t.Run("Renewal", func(t *testing.T) {
		w, _ := NewNodeIDWorker(rdb)
		id, _ := w.Acquire()
		key := WorkerKeyPrefix + fmt.Sprintf("%d", id)
		time.Sleep(RenewInterval + time.Second)

		// 验证租约被续期
		ttl, err := rdb.TTL(ctx, key).Result()
		if err != nil {
			t.Fatalf("Failed to get TTL: %v", err)
		}
		if ttl <= 0 {
			t.Fatalf("Lease not renewed, TTL: %v", ttl)
		}

		w.Release()
	})

	t.Run("RecycleExpired", func(t *testing.T) {
		w, _ := NewNodeIDWorker(rdb)
		id, _ := w.Acquire()

		// 手动删除worker key以模拟过期
		key := WorkerKeyPrefix + fmt.Sprintf("%d", id)
		rdb.Del(ctx, key)

		err := w.recycleExpiredWorkers()
		if err != nil {
			t.Fatalf("Failed to recycle expired workers: %v", err)
		}

		// 验证ID被回收
		exists, err := rdb.SIsMember(ctx, AvailableKey, id).Result()
		if err != nil {
			t.Fatalf("Failed to check if ID is in available set: %v", err)
		}
		if !exists {
			t.Fatalf("ID %d was not released back to the pool", id)
		}
	})
}
