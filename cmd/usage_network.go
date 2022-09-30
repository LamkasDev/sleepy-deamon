package main

import (
	"runtime"
)

type NetworkUsage struct {
	RX int64 `json:"rx"`
	TX int64 `json:"tx"`
}

func GetNetworkUsage() NetworkUsage {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetNetworkUsageSystem()
	default:
		return NetworkUsage{}
	}
}
