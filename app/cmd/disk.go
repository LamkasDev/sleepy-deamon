package main

import (
	"runtime"
)

type Disk struct {
	ID       string      `json:"id"`
	Parent   string      `json:"parent"`
	Name     string      `json:"name"`
	SSD      bool        `json:"ssd"`
	PTUUID   *string     `json:"ptuuid"`
	Size     uint64      `json:"size"`
	Model    *string     `json:"model"`
	Children []Partition `json:"children"`
}

type Partition struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	UUID       *string `json:"uuid"`
	PartUUID   *string `json:"partuuid"`
	Type       *string `json:"type"`
	Size       uint64  `json:"size"`
	Used       *uint64 `json:"used"`
	Mountpoint *string `json:"mountpoint"`
}

func GetDisks() []Disk {
	switch runtime.GOOS {
	case "linux", "windows":
		return GetDisksSystem()
	default:
		return []Disk{}
	}
}
