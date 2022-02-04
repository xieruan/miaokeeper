package memutils

type MemDriverRedis struct {
	MemDriver
}

func (md *MemDriverRedis) Init(kargs ...string) {

}

func (md *MemDriverRedis) Read(key string) (interface{}, bool) {
	return nil, false
}

func (md *MemDriverRedis) Write(key string, value string, expire int64, overwriteTTLIfExists bool) interface{} {
	return nil
}

func (md *MemDriverRedis) Inc(key string, expire int64, overwriteTTLIfExists bool) int {
	return 0
}

func (md *MemDriverRedis) Expire(key string) {

}

func (md *MemDriverRedis) Exists(key string) bool {
	return false
}
