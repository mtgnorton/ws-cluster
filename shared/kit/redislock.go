package kit

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Locker interface {
	Lock(ctx context.Context, key string, expiration time.Duration) error
}

type UnLocker interface {
	Unlock(ctx context.Context) error
}

type RedisLocker struct {
	redisClient *redis.Client
	prefix      string
}

type RedisUnLocker struct {
	redisClient *redis.Client
	k           string
	v           string
}

var (
	ErrLockFailed     = errors.New("lock failed")
	ErrLockerNotExist = errors.New("locker not exist")
)

var luaReleaseScript = redis.NewScript(`
if redis.call("get", KEYS[1]) == ARGV[1] then
    return redis.call("del", KEYS[1])
else
    return 0
end
`)

func (r *RedisLocker) Lock(ctx context.Context, k string, expiration time.Duration) (UnLocker, error) {
	k = r.prefix + k
	v, err := GenerateUniqueId()
	if err != nil {
		return nil, err
	}
	ok, err := r.redisClient.SetNX(ctx, k, v, expiration).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrLockFailed
	}
	return &RedisUnLocker{
		redisClient: r.redisClient,
		k:           k,
		v:           strconv.FormatInt(v, 10),
	}, nil
}

func (r *RedisUnLocker) Unlock(ctx context.Context) error {
	if r == nil {
		return ErrLockerNotExist
	}
	rs, err := luaReleaseScript.Run(ctx, r.redisClient, []string{r.k}, r.v).Result()
	if errors.Is(err, redis.Nil) {
		return ErrLockerNotExist
	} else if err != nil {
		return err
	}
	if v, ok := rs.(int64); !ok || v != 1 {
		return ErrLockerNotExist
	}
	return nil
}

func NewRedisLocker(redisClient *redis.Client, prefix string) *RedisLocker {
	return &RedisLocker{
		redisClient: redisClient,
		prefix:      prefix,
	}
}
