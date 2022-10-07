package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func MathMin(n int64, min int64) int64 {
	if n < min {
		return min
	}

	return n
}

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

func GetDockerFormat(fields []string) string {
	inner := strings.Join(ArrayMap(fields, func(e string) string { return fmt.Sprintf(`"%s":"{{.%s}}"`, e, e) }), ",")
	return fmt.Sprintf("{%s}", inner)
}

func ConvertToBytesShort(raw string) uint64 {
	if len(raw) > 1 {
		part := raw[len(raw)-1:]
		partNum, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
		switch part {
		case "B":
			return uint64(partNum)
		case "K":
			return uint64(partNum * 1024)
		case "M":
			return uint64(partNum * 1048576)
		case "G":
			return uint64(partNum * 1073741824)
		case "T":
			return uint64(partNum * 1099511627776)
		}
	}

	return 0
}

func ConvertToBytes(raw string) uint64 {
	if len(raw) > 3 {
		binPart := raw[len(raw)-3:]
		binPartNum, _ := strconv.ParseFloat(raw[:len(raw)-3], 32)
		switch binPart {
		case "KiB":
			return uint64(binPartNum * 1024)
		case "MiB":
			return uint64(binPartNum * 1048576)
		case "GiB":
			return uint64(binPartNum * 1073741824)
		}
	}
	if len(raw) > 2 {
		decPart := raw[len(raw)-2:]
		decPartNum, _ := strconv.ParseFloat(raw[:len(raw)-2], 32)
		switch decPart {
		case "kB":
			return uint64(decPartNum * 1000)
		case "MB":
			return uint64(decPartNum * 1000000)
		case "GB":
			return uint64(decPartNum * 1000000000)
		}
	}
	if len(raw) > 1 {
		bPart := raw[len(raw)-1:]
		bPartNum, _ := strconv.ParseFloat(raw[:len(raw)-1], 32)
		if bPart == "B" {
			return uint64(bPartNum)
		}
	}

	SleepyWarnLn("Failed to convert %s to bytes!", raw)
	return 0
}

func DirSize(path string) (int64, error) {
	var size int64
	err := filepath.Walk(path, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return err
	})
	return size, err
}
