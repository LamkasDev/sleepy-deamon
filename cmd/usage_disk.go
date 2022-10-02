package main

import "runtime"

type DiskUsage struct {
	Parent       string `json:"parent"`
	Read         uint64 `json:"read"`
	Write        uint64 `json:"write"`
	ReadLatency  uint64 `json:"readLatency"`
	WriteLatency uint64 `json:"writeLatency"`
}

type DiskUsageRaw struct {
	Name         string
	Reads        uint64
	ReadSectors  uint64
	ReadTime     uint64
	Writes       uint64
	WriteSectors uint64
	WriteTime    uint64
}

func GetDiskUsages() []DiskUsageRaw {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetDiskUsagesSystem()
	default:
		return []DiskUsageRaw{}
	}
}
