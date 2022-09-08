//go:build windows
// +build windows

package main

import (
	"unsafe"
)

type MemoryWindowsRaw struct {
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

func GetMemorySystem() (MemoryWindowsRaw, error) {
	var memory MemoryWindowsRaw
	memory.Length = uint32(unsafe.Sizeof(memory))
	ret, _, err := globalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&memory)))
	if ret == 0 {
		return MemoryWindowsRaw{}, err
	}

	return memory, nil
}

func GetMemoryDetailsSystem() (MemoryState, MemoryUsage) {
	memory, err := GetMemorySystem()
	if err != nil {
		SleepyWarnLn("Failed to get memory details! (%s)", err.Error())
		return MemoryState{}, MemoryUsage{}
	}
	return MemoryState{
			Total:     memory.TotalPhys,
			SwapTotal: memory.TotalPageFile,
		}, MemoryUsage{
			Used:     (float32(memory.AvailPhys) / float32(memory.TotalPhys+1)) * 100,
			SwapUsed: (float32(memory.AvailPageFile) / float32(memory.TotalPageFile+1)) * 100,
		}
}
