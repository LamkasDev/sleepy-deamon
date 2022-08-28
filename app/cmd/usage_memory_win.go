//go:build windows
// +build windows

package main

import (
	"unsafe"
)

type MemoryUsageWindowsRaw struct {
	Length               uint32
	MemoryLoad           uint32
	TotalPhys            uint64
	AvailPhys            uint64
	TotalPageFile        uint64
	AvailPageFile        uint64
	TotalVirtual         uint64
	AvailVirtual         uint64
	AvailExtendedVirtual uint64
}

func GetMemoryUsageSystem() MemoryUsage {
	var memory MemoryUsageWindowsRaw
	memory.Length = uint32(unsafe.Sizeof(memory))

	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memory)))
	if ret == 0 {
		SleepyWarnLn("Failed to get RAM usage! (%s)", err.Error())
		return MemoryUsage{}
	}

	return MemoryUsage{
		Total:     memory.TotalPhys,
		Used:      memory.TotalPhys - memory.AvailPhys,
		SwapTotal: memory.TotalPageFile,
		SwapUsed:  memory.TotalPageFile - memory.AvailPageFile,
	}
}
