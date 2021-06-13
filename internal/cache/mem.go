package cache

import (
	"context"
	"time"

	"github.com/go-redis/cache/v8"
)

type Memory struct {
	C   *cache.Cache
	TTL time.Duration
}

func NewMemory(ttl time.Duration) *Memory {
	return &Memory{cache.New(&cache.Options{
		LocalCache: cache.NewTinyLFU(1000, ttl),
	}), ttl}
}

func (m *Memory) Set(key string, val interface{}) error {
	ctx := context.Background()
	return m.C.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
		TTL:   m.TTL,
	})
}

func (m *Memory) Get(key string, value interface{}) error {
	ctx := context.Background()
	return m.C.Get(ctx, key, value)
}

func (m *Memory) Drop(key string) error {
	ctx := context.Background()
	return m.C.Delete(ctx, key)
}
