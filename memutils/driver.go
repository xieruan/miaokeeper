package memutils

import "time"

type MemDriver interface {
	Init(kargs ...string)

	Read(key string) (interface{}, bool)
	Write(key string, value interface{}, expire int64, overwriteTTLIfExists bool) interface{}
	Inc(key string, expire int64, overwriteTTLIfExists bool) int

	Expire(key string)
	Exists(key string) bool
}

func Now() int64 {
	return time.Now().UnixMilli()
}
