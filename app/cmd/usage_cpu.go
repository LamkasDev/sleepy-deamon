package main

import "runtime"

type CPUUsage struct {
	Total float32 `json:"total"`
}

type CPUUsageRaw struct {
	User, Nice, System, Idle, Iowait, Irq, Softirq, Steal, Guest, GuestNice, Total uint64
	CPUCount, StatCount                                                            int
}

func GetCPUUsage() CPUUsageRaw {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetCPUUsageSystem()
	default:
		return CPUUsageRaw{}
	}
}

// TODO: make windows implementation
