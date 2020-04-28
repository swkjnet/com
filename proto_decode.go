//协议反序列化
package com

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"math"
	"reflect"
	"sync"
	"unsafe"
)

func UnPack(data []byte, v interface{}) error {
	d := decodeState{*bytes.NewBuffer(data)}
	return d.unmarshal(v)
}

func (d *decodeState) unmarshal(v interface{}) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("unpack not support type:" + reflect.TypeOf(v).String())
	}
	d.reflectValue(rv.Elem())
	return nil
}

//按长度读取字符串
func (d *decodeState) ReadBytesLen(l uint16) []byte {
	buf := make([]byte, l)
	_, err := io.ReadFull(d, buf)
	if err != nil {
		panic(err)
	}
	return buf
}

//读取数字类型(需要明确v指针类型)
func (d *decodeState) ReadInt(v interface{}) {
	err := binary.Read(d, binary.BigEndian, v)
	if err != nil {
		panic(err)
	}
}

type decodeState struct {
	bytes.Buffer
}

func (d *decodeState) reflectValue(v reflect.Value) {
	typeDecoder(v.Type())(d, v)
}

type decoderFunc func(e *decodeState, v reflect.Value)

var decoderCache sync.Map // map[reflect.Type]decoderFunc

func typeDecoder(t reflect.Type) decoderFunc {
	if fi, ok := decoderCache.Load(t); ok {
		return fi.(decoderFunc)
	}
	var (
		wg sync.WaitGroup
		f  decoderFunc
	)
	wg.Add(1)
	fi, loaded := decoderCache.LoadOrStore(t, decoderFunc(func(e *decodeState, v reflect.Value) {
		wg.Wait()
		f(e, v)
	}))
	if loaded {
		return fi.(decoderFunc)
	}
	f = newTypeDecoder(t)
	wg.Done()
	decoderCache.Store(t, f)
	return f
}
func newTypeDecoder(t reflect.Type) decoderFunc {
	switch t.Kind() {
	case reflect.Bool:
		return boolDecoder
	case reflect.Int:
		return intDecoder
	case reflect.Int32:
		return int32Decoder
	case reflect.Uint:
		return uintDecoder
	case reflect.Uint32:
		return uint32Decoder
	case reflect.Int8:
		return int8Decoder
	case reflect.Int16:
		return int16Decoder
	case reflect.Int64:
		return int64Decoder
	case reflect.Uint8:
		return uint8Decoder
	case reflect.Uint16:
		return uint16Decoder
	case reflect.Uint64:
		return uint64Decoder
	case reflect.Float32:
		return float32Decoder
	case reflect.Float64:
		return float64Decoder
	case reflect.String:
		return stringDecoder
	case reflect.Struct:
		return newStructDecoder(t)
	case reflect.Slice:
		return newSliceDecoder(t)
	case reflect.Array:
		return newArrayDecoder(t)
	default:
		panic(errors.New("proto unpack: unsupported type: " + t.String()))
	}
}

func boolDecoder(d *decodeState, v reflect.Value) {
	b, err := d.ReadByte()
	if err != nil {
		panic(err)
	}
	if b == 1 {
		v.SetBool(true)
	} else {
		v.SetBool(false)
	}
}

//读取四个字节
func intDecoder(d *decodeState, v reflect.Value) {
	var t int32
	d.ReadInt(&t)
	v.SetInt(int64(t))
}

//读取四个字节
func int32Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*int32)((unsafe.Pointer)(v.UnsafeAddr())))
}

//读取四个字节
func uintDecoder(d *decodeState, v reflect.Value) {
	var t uint32
	d.ReadInt(&t)
	v.SetUint(uint64(t))
}
func uint32Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*uint32)((unsafe.Pointer)(v.UnsafeAddr())))
}

func int8Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*int8)((unsafe.Pointer)(v.UnsafeAddr())))
}

func int16Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*int16)((unsafe.Pointer)(v.UnsafeAddr())))
}

func int64Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*int64)((unsafe.Pointer)(v.UnsafeAddr())))
}

func uint8Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*uint8)((unsafe.Pointer)(v.UnsafeAddr())))
}

func uint16Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*uint16)((unsafe.Pointer)(v.UnsafeAddr())))
}

func uint64Decoder(d *decodeState, v reflect.Value) {
	d.ReadInt((*uint64)((unsafe.Pointer)(v.UnsafeAddr())))
}

//float32
func float32Decoder(d *decodeState, v reflect.Value) {
	var b uint32 = 0
	d.ReadInt(&b)
	v.SetFloat(float64(math.Float32frombits(b)))
}

//float64
func float64Decoder(d *decodeState, v reflect.Value) {
	var b uint64 = 0
	d.ReadInt(&b)
	v.SetFloat(math.Float64frombits(b))
}

//string解析
func stringDecoder(d *decodeState, v reflect.Value) {
	var l uint16 = 0
	d.ReadInt(&l)
	if l == 0 {
		return
	}
	b := make([]byte, l)
	_, err := io.ReadFull(d, b)
	if err != nil {
		panic(err)
	}
	v.SetString(string(b))
}

func newSliceDecoder(t reflect.Type) decoderFunc {
	if t.Elem().Kind() == reflect.Uint8 {
		return decodeByteSlice
	}
	dec := sliceDecoder{newArrayDecoder(t)}
	return dec.decode
}

func newArrayDecoder(t reflect.Type) decoderFunc {
	dec := arrayDecoder{typeDecoder(t.Elem())}
	return dec.decode
}

func decodeByteSlice(d *decodeState, v reflect.Value) {
	var l uint16
	d.ReadInt(&l)
	if l == 0 {
		return
	}
	v.SetBytes(d.ReadBytesLen(l))
}

type sliceDecoder struct {
	arrayDec decoderFunc
}

type arrayDecoder struct {
	elemDec decoderFunc
}

func (se sliceDecoder) decode(d *decodeState, v reflect.Value) {
	se.arrayDec(d, v)
}

func (ae arrayDecoder) decode(d *decodeState, v reflect.Value) {
	var l uint16
	d.ReadInt(&l)
	if l == 0 {
		return
	}
	L := int(l)
	if v.Kind() == reflect.Slice {
		v.Set(reflect.MakeSlice(v.Type(), L, L))
	}
	for i := 0; i < L; i++ {
		ae.elemDec(d, v.Index(i))
	}
}

type structDecoder struct {
	fields []deField
}

func (se structDecoder) decode(d *decodeState, v reflect.Value) {
	for i := range se.fields {
		f := &se.fields[i]
		f.decoder(d, v.Field(i))
	}
}

var deFieldCache sync.Map // map[reflect.Type]structDecoder

type deField struct {
	decoder decoderFunc
}

func newStructDecoder(t reflect.Type) decoderFunc {
	return cachedDeFields(t).decode
}

func cachedDeFields(t reflect.Type) structDecoder {
	if f, ok := deFieldCache.Load(t); ok {
		return f.(structDecoder)
	}
	f, _ := deFieldCache.LoadOrStore(t, typeDeFields(t))
	return f.(structDecoder)
}

func typeDeFields(t reflect.Type) structDecoder {
	var fields []deField
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		st := sf.Type
		if st == t {
			panic("proto:not support recursive type:" + st.String())
		}
		if sf.PkgPath != "" {
			panic("proto:not support type:" + t.String() + ",name:" + sf.Name)
		}
		fields = append(fields, deField{decoder: typeDecoder(st)})
	}
	return structDecoder{fields}
}
