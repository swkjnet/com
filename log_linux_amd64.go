//获取linux崩溃日志(可以根据文件最后修改时间，知悉崩溃时间)
package com

import (
	"os"
	"syscall"
	"time"
)

func init() {
	os.MkdirAll("log", 0660)
	logFile, err := os.OpenFile("log/sysdebug.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
	if err != nil {
		log.Println("open sysdebug err:", err.Error())
		return
	}
	// 将进程标准出错重定向至文件，进程崩溃时运行时将向该文件记录协程调用栈信息
	syscall.Dup2(int(logFile.Fd()), int(os.Stderr.Fd()))
}
