package com

import (
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/swkjnet/proto"
)

func TestUUid(t *testing.T) {
	var arr1 []string
	var arr2 [2]string
	var m uint32
	go func() {
		v1 := time.Now()
		for i := 0; i < 20000; i++ {
			arr1 = append(arr1, GetUUid())
		}
		atomic.AddUint32(&m, 1)
		t.Log("cost1:", time.Since(v1))
	}()
	go func() {
		v1 := time.Now()
		for i := 0; i < 2; i++ {
			arr2[i] = GetUUid()
			//arr2 = append(arr2, GetUUid())
		}
		atomic.AddUint32(&m, 1)
		t.Log("cost2:", time.Since(v1))
	}()
	for {
		if m == 2 {
			v2 := time.Now()
			i5 := ArrIsExist(arr2, arr2[0])
			t.Log("one:", time.Since(v2), ",index:", i5)
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func CheckArr(arr []string, s string) int {
	//return ArrIsExist(arr, s)
	for k, v := range arr {
		if v == s {
			return k
		}
	}
	return -1
}

type Ad struct {
	Bool   bool
	Uint   uint
	Int    int
	Int8   int8
	Int16  int16
	Int32  int32
	Int64  int64
	Uint8  uint8
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
	String string
	Slice  []Ad
	Arr    [10]int
}

type BStruct struct {
	Serverid []int
}

func TestMarshal(t *testing.T) {
	return
	//bs := BStruct{Serverid: []int{1}}
	buf := []byte{0, 1, 0, 0, 0, 1}
	bs := &BStruct{}
	err3 := UnPack(buf, bs)
	if err3 != nil {
		t.Log("err:", err3.Error())
		return
	}
	t.Log(bs)
	return
	var arr *Ad = &Ad{
		Bool:   true,
		Uint:   3999999999,
		Int:    -3,
		Int8:   -4,
		Int16:  -5,
		Int32:  -6,
		Int64:  -7,
		Uint8:  8,
		Uint16: 9,
		Uint32: 10,
		Uint64: 11,
		String: "string",
		Arr:    [10]int{1, 2, -3, -4, -5, -6, -7, 8, 9, 10},
	}
	v := uint8(64)
	json.Marshal(v)
	res, err1 := Pack(arr)
	if err1 != nil {
		t.Log(err1.Error())
		return
	}
	t.Log("marshal:", res)
	var bb Ad
	err := UnPack(res, &bb)
	if err != nil {
		t.Log(err.Error())
		return
	}
	t.Log("unmarshal:", bb)
	res, _ = proto.Pack(arr)
	t.Log("proto:", res)
	// arr := int32(30)
	// // for i := 0; i < 10000; i++ {
	// // 	arr = append(arr, GetUUid())
	// // }
	// v1 := time.Now()
	// //v := ArrIsExist(arr, "1")
	// v, err := Marshal(arr)
	// if err != nil {
	// 	t.Log("error:", err.Error())
	// 	return
	// }
	// t.Log(v, ",cost:", int64(time.Since(v1)))
	// v1 = time.Now()
	// var arr1 *Ad = &Ad{"9999991", 1, 25}
	// v2, _ := Marshal(arr1)
	// t.Log(v2, ",cost1:", int64(time.Since(v1)))
}
