package memutils

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
)

type MemDriverRedis struct {
	MemDriver

	rdb *redis.Client
	ctx context.Context
}

func (md *MemDriverRedis) Init(kargs ...string) {
	if len(kargs) < 2 {
		fmt.Fprintln(os.Stderr, "MemDriver Error | Should have two parameters to indicate host and password")
		os.Exit(1)
	}

	md.ctx = context.Background()
	md.rdb = redis.NewClient(&redis.Options{
		Addr:     kargs[0],
		Password: kargs[1],
		DB:       0,
	})
}

func (md *MemDriverRedis) Read(key string) (interface{}, bool) {
	val, err := md.rdb.Get(md.ctx, key).Result()
	return val, err == nil
}

func (md *MemDriverRedis) Write(key string, value interface{}, expire int64, overwriteTTLIfExists bool) interface{} {
	expireAfter := time.Duration(expire * int64(time.Millisecond))
	if !overwriteTTLIfExists {
		if duration, err := md.rdb.TTL(md.ctx, key).Result(); err == nil {
			expireAfter = duration
		}
	}
	md.rdb.Set(md.ctx, key, value, expireAfter)

	return value
}

func (md *MemDriverRedis) Inc(key string, expire int64, overwriteTTLIfExists bool) int {
	expireAfter := time.Duration(expire * int64(time.Millisecond))
	if !overwriteTTLIfExists {
		if duration, err := md.rdb.TTL(md.ctx, key).Result(); err == nil {
			expireAfter = duration
		}
	}
	ret := md.rdb.Incr(md.ctx, key).Val()
	md.rdb.Expire(md.ctx, key, expireAfter)

	return int(ret)
}

func (md *MemDriverRedis) Expire(key string) {
	md.rdb.Unlink(md.ctx, key)
}

func (md *MemDriverRedis) Exists(key string) bool {
	return md.rdb.Exists(md.ctx, key).Val() == 1
}
