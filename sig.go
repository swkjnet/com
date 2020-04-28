//sha1、md5签名
package com

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
)

/*
sha1签名后进行base64编码
@param 源串
@param 密钥
@return 签名结果
*/
func Base64_sha1(val string, key string) string {
	return Base64String(Hmac_sha1(val, key))
}

/*
base64编码
@param 源串
@return 编码结果
*/
func Base64String(str []byte) string {
	return base64.StdEncoding.EncodeToString(str)
}

/*
sha1签名
@param 源串
@param 密钥
@return 签名结果
*/
func Hmac_sha1(val string, key string) []byte {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(val))
	return h.Sum(nil)
}

/*
md5签名
*/
func Md5Str(val string) string {
	h := md5.New()
	h.Write([]byte(val))
	return fmt.Sprintf("%x", h.Sum(nil))
}
