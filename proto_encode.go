//协议序列化
package com

import (
	"bytes"
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"sync"
)

func Pack(v interface{}) ([]byte, error) {
	e := newEncodeState()
	err := e.marshal(v)
	if err != nil {
		return nil, err
	}
	buf := append([]byte(nil), e.Bytes()...)
	encodeStatePool.Put(e)
	return buf, nil
}

var encodeStatePool sync.Pool

type encodeState struct {
	bytes.Buffer
}

//写入数字类型（v需要明确数字类型，不能为int，uint）
func (e *encodeState) WriteInt(v interface{}) {
	err := binary.Write(e, binary.BigEndian, v)
	if err != nil {
		panic(err)
	}
}

func (e *encodeState) reflectValue(v reflect.Value) {
	typeEncoder(v.Type())(e, v)
}

func (e *encodeState) marshal(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if je, ok := r.(error); ok {
				err = je
			} else {
				panic(r)
			}
		}
	}()
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	e.reflectValue(rv)
	return
}

func newEncodeState() *encodeState {
	if v := encodeStatePool.Get(); v != nil {
		e := v.(*encodeState)
		e.Reset()
		return e
	}
	return &encodeState{}
}

type protoError struct{ error }

type encoderFunc func(e *encodeState, v reflect.Value)

var encoderCache sync.Map // map[reflect.Type]encoderFunc

func typeEncoder(t reflect.Type) encoderFunc {
	if fi, ok := encoderCache.Load(t); ok {
		return fi.(encoderFunc)
	}
	var (
		wg sync.WaitGroup
		f  encoderFunc
	)
	wg.Add(1)
	fi, loaded := encoderCache.LoadOrStore(t, encoderFunc(func(e *encodeState, v reflect.Value) {
		wg.Wait()
		f(e, v)
	}))
	if loaded {
		return fi.(encoderFunc)
	}
	f = newTypeEncoder(t)
	wg.Done()
	encoderCache.Store(t, f)
	return f
}

func newTypeEncoder(t reflect.Type) encoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolEncoder
	case reflect.Int:
		return intEncoder
	case reflect.Uint:
		return uintEncoder
	case reflect.Int32, reflect.Int8, reflect.Int16, reflect.Int64, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return intComEncoder
	case reflect.Float32:
		return float32Encoder
	case reflect.Float64:
		return float64Encoder
	case reflect.String:
		return stringEncoder
	case reflect.Struct:
		return newStructEncoder(t)
	case reflect.Slice:
		return newSliceEncoder(t)
	case reflect.Array:
		return newArrayEncoder(t)
	default:
		panic(errors.New("proto: unsupported type: " + t.String()))
	}
}

func boolEncoder(e *encodeState, v reflect.Value) {
	if v.Bool() {
		e.WriteByte(1)
	} else {
		e.WriteByte(0)
	}
}

//通用
func intComEncoder(e *encodeState, v reflect.Value) {
	e.WriteInt(v.Interface())
}

//4字节
func intEncoder(e *encodeState, v reflect.Value) {
	e.WriteInt(int32(v.Int()))
}

//4字节
func uintEncoder(e *encodeState, v reflect.Value) {
	e.WriteInt(uint32(v.Uint()))
}

//float32
func float32Encoder(e *encodeState, v reflect.Value) {
	e.WriteInt(math.Float32bits((float32)(v.Float())))
}

//float64
func float64Encoder(e *encodeState, v reflect.Value) {
	e.WriteInt(math.Float64bits((v.Float())))
}

//string
func stringEncoder(e *encodeState, v reflect.Value) {
	buf := []byte(v.String())
	l := len(buf)
	e.WriteInt(uint16(l))
	e.Write(buf)
}

func encodeByteSlice(e *encodeState, v reflect.Value) {
	res := v.Bytes()
	l := len(res)
	if l >= 65535 {
		panic(errors.New("proto: []uint8 len > 65535"))
	}
	e.WriteInt(uint16(l))
	e.Write(res)
}

type sliceEncoder struct {
	arrayEnc encoderFunc
}

func (se sliceEncoder) encode(e *encodeState, v reflect.Value) {
	if v.IsNil() {
		e.WriteInt(uint16(0))
		return
	}
	se.arrayEnc(e, v)
}

func newSliceEncoder(t reflect.Type) encoderFunc {
	// Byte slices get special treatment; arrays don't.
	if t.Elem().Kind() == reflect.Uint8 {
		return encodeByteSlice
	}
	enc := sliceEncoder{newArrayEncoder(t)}
	return enc.encode
}

type arrayEncoder struct {
	elemEnc encoderFunc
}

func (ae arrayEncoder) encode(e *encodeState, v reflect.Value) {
	l := v.Len()
	if l > 65535 {
		panic(errors.New("proto: not support arrLen> 65535 type:" + v.Type().String()))
	}
	e.WriteInt(uint16(l))
	for i := 0; i < l; i++ {
		ae.elemEnc(e, v.Index(i))
	}
}

func newArrayEncoder(t reflect.Type) encoderFunc {
	enc := arrayEncoder{typeEncoder(t.Elem())}
	return enc.encode
}

type structEncoder struct {
	fields []field
}

func (se structEncoder) encode(e *encodeState, v reflect.Value) {
	for i := range se.fields {
		f := &se.fields[i]
		f.encoder(e, v.Field(i))
	}
}

func newStructEncoder(t reflect.Type) encoderFunc {
	return cachedTypeFields(t).encode
}

type field struct {
	encoder encoderFunc
}

func typeFields(t reflect.Type) structEncoder {
	var fields []field
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		st := sf.Type
		if st == t {
			panic("proto:not support recursive type:" + st.String())
		}
		if sf.PkgPath != "" {
			panic("proto:not support type:" + t.String() + ",name:" + sf.Name)
		}
		fields = append(fields, field{encoder: typeEncoder(st)})
	}
	return structEncoder{fields}
}

var fieldCache sync.Map // map[reflect.Type]structEncoder

func cachedTypeFields(t reflect.Type) structEncoder {
	if f, ok := fieldCache.Load(t); ok {
		return f.(structEncoder)
	}
	f, _ := fieldCache.LoadOrStore(t, typeFields(t))
	return f.(structEncoder)
}
