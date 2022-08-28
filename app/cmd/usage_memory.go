package main

import (
	"runtime"
)

type MemoryUsage struct {
	Total     uint64 `json:"total"`
	Used      uint64 `json:"used"`
	SwapTotal uint64 `json:"swapTotal"`
	SwapUsed  uint64 `json:"swapUsed"`
}

func GetMemoryUsage() MemoryUsage {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetMemoryUsageSystem()
	default:
		return MemoryUsage{}
	}
}
