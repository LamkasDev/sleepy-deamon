package main

import (
	"runtime"
)

type NetworkUsage struct {
	RX uint64 `json:"rx"`
	TX uint64 `json:"tx"`
}

func GetNetworkUsage() NetworkUsage {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetNetworkUsageSystem()
	default:
		return NetworkUsage{}
	}
}
