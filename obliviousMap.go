package main

import (
	"sync"
	"time"
)

type ObliviousMapIfce struct {
	duration int64
	timer    map[string]int64
	inner    map[string]interface{}

	utif bool

	lock sync.Mutex
}

func NowTS() int64 {
	return time.Now().UnixMilli()
}

func (om *ObliviousMapIfce) unsafeGet(key string) (interface{}, bool) {
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

func (om *ObliviousMapIfce) Get(key string) (interface{}, bool) {
	om.lock.Lock()
	defer om.lock.Unlock()

	return om.unsafeGet(key)
}

func (om *ObliviousMapIfce) unsafeSet(key string, value interface{}) interface{} {
	now := NowTS()
	_, ok := om.inner[key]
	om.inner[key] = value
	if !ok || om.utif {
		om.timer[key] = now + om.duration
	}
	return value
}

func (om *ObliviousMapIfce) Set(key string, value interface{}) interface{} {
	om.lock.Lock()
	defer om.lock.Unlock()

	return om.unsafeSet(key, value)
}

func (om *ObliviousMapIfce) Unset(key string) {
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

func (om *ObliviousMapIfce) Exist(key string) bool {
	om.lock.Lock()
	defer om.lock.Unlock()

	_, ok := om.inner[key]
	return ok
}

type ObliviousMapInt struct {
	*ObliviousMapIfce
}

func (om *ObliviousMapInt) Add(key string) int {
	om.lock.Lock()
	defer om.lock.Unlock()

	v, _ := om.unsafeGet(key)
	return om.unsafeSet(key, v.(int)+1).(int)
}

func (om *ObliviousMapInt) Get(key string) (int, bool) {
	a, b := om.ObliviousMapIfce.Get(key)
	return a.(int), b
}

func (om *ObliviousMapInt) Set(key string, value int) int {
	return om.ObliviousMapIfce.Set(key, value).(int)
}

type ObliviousMapStr struct {
	*ObliviousMapIfce
}

func (om *ObliviousMapStr) Get(key string) (string, bool) {
	a, b := om.ObliviousMapIfce.Get(key)
	return a.(string), b
}

func (om *ObliviousMapStr) Set(key string, value string) string {
	return om.ObliviousMapIfce.Set(key, value).(string)
}

func NewOMapIfce(duration int64, updateTimeIfWrite bool) *ObliviousMapIfce {
	return &ObliviousMapIfce{
		duration: duration,
		timer:    make(map[string]int64),
		inner:    make(map[string]interface{}),
		utif:     updateTimeIfWrite,
	}
}

func NewOMapInt(duration int64, updateTimeIfWrite bool) *ObliviousMapInt {
	return &ObliviousMapInt{
		ObliviousMapIfce: NewOMapIfce(duration, updateTimeIfWrite),
	}
}

func NewOMapStr(duration int64, updateTimeIfWrite bool) *ObliviousMapStr {
	return &ObliviousMapStr{
		ObliviousMapIfce: NewOMapIfce(duration, updateTimeIfWrite),
	}
}
