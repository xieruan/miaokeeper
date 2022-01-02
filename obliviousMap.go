package main

import (
	"sync"
	"time"
)

type ObliviousMap struct {
	duration int64
	timer    map[string]int64
	inner    map[string]int

	utif bool

	lock sync.Mutex
}

func NowTS() int64 {
	return time.Now().UnixMilli()
}

func NewOMap(duration int64, updateTimeIfWrite bool) *ObliviousMap {
	return &ObliviousMap{
		duration: duration,
		timer:    make(map[string]int64),
		inner:    make(map[string]int),
		utif:     updateTimeIfWrite,
	}
}

func (om *ObliviousMap) unsafeGet(key string) (int, bool) {
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

func (om *ObliviousMap) Get(key string) (int, bool) {
	om.lock.Lock()
	defer om.lock.Unlock()

	return om.unsafeGet(key)
}

func (om *ObliviousMap) unsafeSet(key string, value int) int {
	now := NowTS()
	_, ok := om.inner[key]
	om.inner[key] = value
	if !ok || om.utif {
		om.timer[key] = now + om.duration
	}
	return value
}

func (om *ObliviousMap) Set(key string, value int) int {
	om.lock.Lock()
	defer om.lock.Unlock()

	return om.unsafeSet(key, value)
}

func (om *ObliviousMap) Unset(key string) {
	om.lock.Lock()
	defer om.lock.Unlock()

	_, ok := om.inner[key]
	if ok {
		delete(om.inner, key)
	}
	_, ok = om.timer[key]
	if ok {
		delete(om.timer, key)
	}
}

func (om *ObliviousMap) Exist(key string) bool {
	om.lock.Lock()
	defer om.lock.Unlock()

	_, ok := om.inner[key]
	return ok
}

func (om *ObliviousMap) Add(key string) int {
	om.lock.Lock()
	defer om.lock.Unlock()

	v, _ := om.unsafeGet(key)
	return om.unsafeSet(key, v+1)
}
