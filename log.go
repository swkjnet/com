//日志
package com

import (
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)

var (
	IsLinux    bool   = (runtime.GOOS == "linux") //系统判定
	milliSec   int64  = getMilliSec()             //毫秒数
	zoneOffset int64  = 0                         //时区偏移(秒)
	gH_Log     *H_Log                             //日志句柄
)

//日志句柄
type H_Log struct {
	gFile     *os.File //日志句柄
	t         int64    //毫秒数(更新日志句柄)
	errFile   *os.File //错误日志句柄
	debugFile *os.File //debug日志句柄
}

//判定是不是linux系统
func init() {
	os.MkdirAll("log", 0660)
	_, v := time.Now().Zone()
	zoneOffset = int64(v)
	gH_Log = InitHLog()
	go func() {
		time.Sleep(time.Millisecond)
		atomic.StoreInt64(&milliSec, getMilliSec())
	}()
}

//获取毫秒数
func getMilliSec() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

//获取当前时间毫秒数
func GetMilliSec() int64 {
	return atomic.LoadInt64(&milliSec)
}

//获取时间的字符串格式
func GetDayFormat() string {
	return time.Unix(GetMilliSec()/1000, 0).Format("2006-01-02")
}

//写日志
func (this *H_Log) printLog(v ...interface{}) {
	if !IsLinux {
		log.Println(v...)
	}
	this.refreshDayLog()
	logger := log.New(this.gFile, "", log.LstdFlags)
	logger.Println(v...)
}

//初始化日志句柄
func InitHLog() *H_Log {
	this := new(H_Log)
	err := this.refreshDayLog()
	if err != nil {
		panic(err)
	}
	this.errFile, err = getfile("error")
	this.debugFile, err = getfile("debug")
	if err != nil {
		panic(err)
	}
	return this
}

//刷新日志句柄
func (this *H_Log) refreshDayLog() error {
	var err error
	if GetMilliSec() >= this.t {
		if this.gFile != nil {
			this.gFile.Close()
		}
		this.gFile, err = getfile(GetDayFormat())
		if err != nil {
			return err
		}
		this.t = GetNextHourStamp(0)
	}
	return err
}

//获取文件句柄
func getfile(filename string) (*os.File, error) {
	return os.OpenFile("./log"+filename+".log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0660)
}

//获取某天整点时间戳
func GetCurHourStamp(h int) int64 {
	return ((GetMilliSec()/1000+zoneOffset-int64(h)*3600)/86400*86400 + int64(h)*3600 - zoneOffset) * 1000
}

//获取下一天某个整点时间戳
func GetNextHourStamp(h int) int64 {
	return GetCurHourStamp(h) + 86400000
}

//基础日志
func LogInfo(v ...interface{}) {
	v = append([]interface{}{"[info]"}, v...)
	gH_Log.printLog(v...)
}

//警告日志
func LogWarn(v ...interface{}) {
	v = append([]interface{}{"[warn]"}, v...)
	gH_Log.printLog(v...)
}

//错误日志
func LogError(v ...interface{}) {
	v = append([]interface{}{"[error]"}, v...)
	gH_Log.printLog(v...)
	logger := log.New(gH_Log.errFile, "", log.LstdFlags)
	logger.Println(v...)
}

//协程崩溃日志
func LogDebug(v ...interface{}) {
	v = append([]interface{}{"[error]"}, v...)
	gH_Log.printLog(v...)
	logger := log.New(gH_Log.debugFile, "", log.LstdFlags)
	logger.Println(v...)
}
