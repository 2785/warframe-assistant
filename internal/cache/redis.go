package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
)

type Redis struct {
	C   *cache.Cache
	TTL time.Duration
}

func NewRedis(dsn string, ttl time.Duration) (*Redis, error) {
	opt, err := redis.ParseURL(dsn)
	if err != nil {
		return nil, err
	}

	rdb := redis.NewClient(opt)

	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		return nil, err
	}

	rCache := cache.New(&cache.Options{
		Redis:      rdb,
		LocalCache: cache.NewTinyLFU(1000, time.Minute),
	})

	return &Redis{C: rCache, TTL: ttl}, nil
}

func (r *Redis) Set(key string, val interface{}) error {
	ctx := context.Background()
	return r.C.Set(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: val,
		TTL:   r.TTL,
	})
}

func (r *Redis) Get(key string, val interface{}) error {
	ctx := context.Background()
	err := r.C.Get(ctx, key, val)
	if errors.Is(err, cache.ErrCacheMiss) {
		return &ErrNoRecord{}
	}
	return err
}

func (r *Redis) Once(key string, recv interface{}, do func() (interface{}, error)) error {
	ctx := context.Background()
	return r.C.Once(&cache.Item{
		Ctx:   ctx,
		Key:   key,
		Value: recv,
		TTL:   r.TTL,
		Do: func(*cache.Item) (interface{}, error) {
			return do()
		},
	})
}

func (r *Redis) Drop(key string) error {
	ctx := context.Background()
	err := r.C.Delete(ctx, key)
	if errors.Is(err, cache.ErrCacheMiss) {
		return &ErrNoRecord{}
	}
	return err
}
