//基础结构
package com

import (
	"bytes"
	"compress/zlib"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
)

/*
 压缩
 @param src 源
 @return 加密结果
*/
func DoZlibCompress(src []byte) []byte {
	var in bytes.Buffer
	w := zlib.NewWriter(&in)
	w.Write(src)
	w.Close()
	return in.Bytes()
}

/*
 获取JSON配置文件
*/
func GetJsonConfig(strfile string, v interface{}) error {
	file, err := os.OpenFile(strfile, os.O_RDWR, 0660)
	if err != nil {
		return err
	}
	buf, err1 := ioutil.ReadAll(file)
	if err1 != nil {
		return err1
	}
	file.Close()
	jerr := json.Unmarshal(buf, v)
	if jerr != nil {
		return jerr
	}
	return nil
}

//判定数组中是否存在某个值
func ArrIsExist(array interface{}, value interface{}) int {
	switch array.(type) {
	case []int:
		arr := array.([]int)
		for k, v := range arr {
			if v == value {
				return k
			}
		}
	case []string:
		arr := array.([]string)
		for k, v := range arr {
			if v == value {
				return k
			}
		}
	default: //性能低
		switch reflect.TypeOf(array).Kind() {
		case reflect.Slice, reflect.Array:
			s := reflect.ValueOf(array)
			for i := 0; i < s.Len(); i++ {
				if reflect.DeepEqual(value, s.Index(i).Interface()) {
					return i
				}
			}
		}
	}
	return -1
}

//生成唯一uuid
func GetUUid() string {
	var v [16]byte
	rand.Read(v[:])
	s := fmt.Sprintf("%x", v)
	return s
}
