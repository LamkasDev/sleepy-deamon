//go:build windows
// +build windows

package main

import "syscall"

var (
	kernel32             = syscall.NewLazyDLL("kernel32.dll")
	globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
	getSystemTimes       = kernel32.NewProc("GetSystemTimes")
)

type FileTimeWindows struct {
	Low  uint32
	High uint32
}
