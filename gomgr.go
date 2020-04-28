//协程管理
package com

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"
)

var gonum int64 = 0 //协程数管理

//开启一个协程 mark协程标识
func StartGo(mark string, f func()) {
	if f == nil {
		return
	}
	atomic.AddInt64(&gonum, 1)
	go func() {
		defer func() {
			if err := recover(); err != nil {
				LogDebug(fmt.Sprint("[debug] ", mark, " error:", err, " stack:", string(debug.Stack())))
			}
			atomic.AddInt64(&gonum, -1)
		}()
		f()
	}()
}
