//go:build windows
// +build windows

package main

import "unsafe"

func GetCPUUsageSystem() CPUUsageRaw {
	var idleTime uint64
	var kernelTime uint64
	var userTime uint64

	ret, _, err := getSystemTimes.Call(uintptr(unsafe.Pointer(&idleTime)), uintptr(unsafe.Pointer(&kernelTime)), uintptr(unsafe.Pointer(&userTime)))
	if ret == 0 {
		SleepyWarnLn("Failed to get CPU usage! (%s)", err.Error())
		return CPUUsageRaw{}
	}

	return CPUUsageRaw{
		Total:  kernelTime + userTime,
		System: kernelTime - idleTime,
		User:   userTime,
	}
}
