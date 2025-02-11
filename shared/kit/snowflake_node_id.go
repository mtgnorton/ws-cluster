package kit

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	AvailableKey    = "snowflake:worker_ids:available"
	WorkerKeyPrefix = "snowflake:worker:"
)
const (
	WorkerTTL     = 30 * time.Second
	RenewInterval = 10 * time.Second
	MaxWorkerID   = 1023
)

var (
	ErrNoAvailableID   = errors.New("no available worker ID")
	ErrAlreadyAcquired = errors.New("worker ID already acquired")
	ErrRedisConn       = errors.New("redis connection error")
	ErrInvalidState    = errors.New("worker in invalid state")
)

type NodeIDWorker struct {
	rdb      *redis.Client
	workerID int64
	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	initOnce sync.Once
}

func NewNodeIDWorker(rdb *redis.Client) (*NodeIDWorker, error) {
	ctx, cancel := context.WithCancel(context.Background())

	w := &NodeIDWorker{
		rdb:      rdb,
		ctx:      ctx,
		cancel:   cancel,
		workerID: -1,
	}

	if err := w.initializePool(); err != nil {
		return nil, WrapError(err, ErrRedisConn)
	}
	return w, nil
}

func (w *NodeIDWorker) initializePool() error {
	var initErr error
	w.initOnce.Do(func() {
		exists, err := w.rdb.Exists(w.ctx, AvailableKey).Result()
		if err != nil {
			initErr = err
			return
		}

		if exists == 0 {
			members := make([]interface{}, MaxWorkerID+1)
			for i := int64(0); i <= MaxWorkerID; i++ {
				members[i] = i
			}
			if err := w.rdb.SAdd(w.ctx, AvailableKey, members...).Err(); err != nil {
				initErr = err
			}
		}
	})
	return initErr
}

func (w *NodeIDWorker) Acquire() (int64, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.workerID != -1 {
		return -1, ErrAlreadyAcquired
	}

	for {
		id, err := w.rdb.SPop(w.ctx, AvailableKey).Int64()
		if err == redis.Nil {
			if err := w.recycleExpiredWorkers(); err != nil {
				return -1, err
			}
			continue
		}
		if err != nil {
			return -1, WrapError(err, ErrRedisConn)
		}

		key := fmt.Sprintf("%s%d", WorkerKeyPrefix, id)
		if ok, err := w.rdb.SetNX(w.ctx, key, "occupied", WorkerTTL).Result(); err != nil {
			w.rdb.SAdd(w.ctx, AvailableKey, id)
			return -1, err
		} else if ok {
			w.workerID = id
			go w.startRenewal(key)
			return id, nil
		}

		w.rdb.SAdd(w.ctx, AvailableKey, id)
	}
}

func (w *NodeIDWorker) Release() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.workerID == -1 {
		return nil
	}

	key := fmt.Sprintf("%s%d", WorkerKeyPrefix, w.workerID)
	if err := w.rdb.Del(w.ctx, key).Err(); err != nil {
		return WrapError(err, ErrRedisConn)
	}

	if err := w.rdb.SAdd(w.ctx, AvailableKey, w.workerID).Err(); err != nil {
		return WrapError(err, ErrRedisConn)
	}

	w.cancel()
	w.workerID = -1
	return nil
}

func (w *NodeIDWorker) startRenewal(key string) {
	ticker := time.NewTicker(RenewInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.mu.Lock()
			if w.workerID == -1 {
				w.mu.Unlock()
				return
			}
			if err := w.rdb.Expire(w.ctx, key, WorkerTTL).Err(); err != nil {
				// 自动处理续期失败
				w.rdb.SAdd(w.ctx, AvailableKey, w.workerID)
				w.workerID = -1
				w.mu.Unlock()
				return
			}
			w.mu.Unlock()
		case <-w.ctx.Done():
			return
		}
	}
}

func (w *NodeIDWorker) recycleExpiredWorkers() error {
	script := `
    for i=0,tonumber(ARGV[1]) do
        local key = ARGV[2] .. i
        if redis.call("EXISTS", key) == 0 then
            redis.call("SADD", KEYS[1], i)
        end
    end
    return redis.call("SCARD", KEYS[1])
    `

	_, err := w.rdb.Eval(w.ctx, script, []string{AvailableKey},
		MaxWorkerID, WorkerKeyPrefix,
	).Result()

	if err != nil {
		return WrapError(err, ErrRedisConn)
	}
	return nil
}
