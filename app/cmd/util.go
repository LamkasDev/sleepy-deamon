package main

import (
	"crypto/md5"
	"encoding/hex"
)

func ArrayMap[I any, O any, F func(I) O](array []I, mapFunc F) []O {
	res := []O{}
	for _, e := range array {
		res = append(res, mapFunc(e))
	}

	return res
}

func GetMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}
