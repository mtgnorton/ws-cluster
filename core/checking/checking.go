package checking

import (
	"time"

	"github.com/sasha-s/go-deadlock"
)

var DefaultChecking *Checking = NewChecking()

type Checking struct {
	pids map[string]struct{}
	mu   deadlock.RWMutex
	opts *Options
}

func NewChecking(opts ...Option) *Checking {
	opt := NewOptions(opts...)
	c := &Checking{opts: opt}
	go c.infiniteGetByRedis()
	return c
}

func (c *Checking) IsExist(pid string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.pids[pid]
	return ok
}

func (c *Checking) infiniteGetByRedis() {
	for {
		c.getPids()
		time.Sleep(c.opts.Interval)
	}
}

func (c *Checking) getPids() error {
	keys, err := c.opts.Redis.HKeys(c.opts.Ctx, "projects").Result()
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.pids = make(map[string]struct{})
	for _, key := range keys {
		c.pids[key] = struct{}{}
	}
	c.mu.Unlock()
	return nil
}
