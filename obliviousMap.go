package main

import (
	"sync"
	"time"
)

type ObliviousMap struct {
	duration int64
	timer    map[string]int64
	inner    map[string]int

	lock sync.Mutex
}

func NowTS() int64 {
	return time.Now().UnixMilli()
}

func NewOMap(duration int64) *ObliviousMap {
	return &ObliviousMap{
		duration: duration,
		timer:    make(map[string]int64),
		inner:    make(map[string]int),
	}
}

func (om *ObliviousMap) Get(key string) (int, bool) {
	om.lock.Lock()
	defer om.lock.Unlock()

	now := NowTS()
	v, ok := om.inner[key]
	if ok {
		if t, ok := om.timer[key]; ok {
			if t <= now {
				delete(om.inner, key)
				delete(om.timer, key)
				return 0, false
			}
			return v, true
		} else {
			delete(om.inner, key)
			return 0, false
		}
	}
	return 0, false
}

func (om *ObliviousMap) Set(key string, value int) int {
	om.lock.Lock()
	defer om.lock.Unlock()

	now := NowTS()
	om.inner[key] = value
	om.timer[key] = now + om.duration
	return value
}
