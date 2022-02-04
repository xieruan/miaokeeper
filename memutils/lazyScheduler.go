package memutils

import (
	"fmt"
	"time"

	jsoniter "github.com/json-iterator/go"
)

type LazySchedulerHandler = func(*LazySchedulerCall)
type LazyScheduler struct {
	driver MemDriver
	fn     map[string]LazySchedulerHandler
}

type LazySchedulerCall struct {
	Name string
	Args map[string]interface{}
}

func (ls *LazyScheduler) Reg(fnName string, handler LazySchedulerHandler) {
	if ls != nil {
		ls.fn[fnName] = handler
	}
}

func (ls *LazyScheduler) ToKey(timestamp int64) string {
	return fmt.Sprintf("ls:%d", timestamp)
}

func (ls *LazyScheduler) After(duration time.Duration, call *LazySchedulerCall) {
	if ls != nil {
		until := Now() + duration.Milliseconds()

		caller, _ := jsoniter.MarshalToString(call)
		ls.driver.Write(ls.ToKey(until), caller, until, true)
	}
}

func (ls *LazyScheduler) Exec(call *LazySchedulerCall) (err error) {
	defer func() {
		if erro := recover(); erro != nil {
			switch x := erro.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown error")
			}
		}
	}()

	if ls != nil && call != nil {
		if fn, ok := ls.fn[call.Name]; ok && fn != nil {
			fn(call)
		}
	}

	return
}

func InitLazyScheduler() {

}
