package main

import (
	"runtime"
)

type MemoryState struct {
	Total     uint64 `json:"total"`
	SwapTotal uint64 `json:"swapTotal"`
}

type MemoryUsage struct {
	Used     float32 `json:"used"`
	SwapUsed float32 `json:"swapUsed"`
}

func GetMemoryDetails() (MemoryState, MemoryUsage) {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetMemoryDetailsSystem()
	default:
		return MemoryState{}, MemoryUsage{}
	}
}
