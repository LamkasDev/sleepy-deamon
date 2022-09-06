package main

import "runtime"

type ZFSPoolRaw struct {
	Name string
	Size string
	Used string
}

type ZFSPool struct {
	ID            string         `json:"id"`
	Name          string         `json:"name"`
	Size          uint64         `json:"size"`
	Used          uint64         `json:"used"`
	Compression   *string        `json:"compression"`
	CompressRatio float32        `json:"compressRatio"`
	Encryption    bool           `json:"encryption"`
	ATime         bool           `json:"atime"`
	Version       uint16         `json:"version"`
	Deduplication bool           `json:"deduplication"`
	RelATime      bool           `json:"relatime"`
	Children      []ZFSPartition `json:"children"`
}

type ZFSPartition struct {
	ID   string `json:"id"`
	Size uint64 `json:"size"`
	Used uint64 `json:"used"`
}

// TODO: add an installed check
func GetZFSPools(disks []Disk) []ZFSPool {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetZFSPoolsSystem(disks)
	default:
		return []ZFSPool{}
	}
}
