package main

import (
	"github.com/BBAlliance/miaokeeper/memutils"
)

type ObliviousMapIfce struct {
	prefix string
	driver memutils.MemDriver

	expire int64
	utif   bool
}

func (om *ObliviousMapIfce) Get(key string) (interface{}, bool) {
	return om.driver.Read(key)
}

func (om *ObliviousMapIfce) Set(key string, value interface{}) interface{} {
	return om.driver.Write(key, value, om.expire, om.utif)
}

func (om *ObliviousMapIfce) Unset(key string) {
	om.driver.Expire(key)
}

func (om *ObliviousMapIfce) Exist(key string) bool {
	return om.driver.Exists(key)
}

type ObliviousMapInt struct {
	*ObliviousMapIfce
}

func (om *ObliviousMapInt) Add(key string) int {
	return om.driver.Inc(key, om.expire, om.utif)
}

func (om *ObliviousMapInt) Get(key string) (int, bool) {
	a, b := om.ObliviousMapIfce.Get(key)
	if a == nil {
		return 0, b
	}
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

func NewOMapIfce(expire int64, updateTimeIfWrite bool, driver memutils.MemDriver) *ObliviousMapIfce {
	return &ObliviousMapIfce{
		expire: expire,
		driver: driver,
		utif:   updateTimeIfWrite,
	}
}

func NewOMapInt(duration int64, updateTimeIfWrite bool, driver memutils.MemDriver) *ObliviousMapInt {
	return &ObliviousMapInt{
		ObliviousMapIfce: NewOMapIfce(duration, updateTimeIfWrite, driver),
	}
}

func NewOMapStr(duration int64, updateTimeIfWrite bool, driver memutils.MemDriver) *ObliviousMapStr {
	return &ObliviousMapStr{
		ObliviousMapIfce: NewOMapIfce(duration, updateTimeIfWrite, driver),
	}
}
