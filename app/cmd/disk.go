package main

import (
	"encoding/json"
	"os/exec"
	"runtime"
)

type DisksRaw struct {
	Blockdevices []DiskRaw
}

type DiskRaw struct {
	Type       string
	PTUUID     *string
	Name       string
	Rota       bool
	Size       uint64
	Model      *string
	UUID       *string
	PartUUID   *string
	FSType     *string
	FSSize     *uint64
	FSUsed     *uint64
	Mountpoint *string
	Children   []PartitionRaw
}

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

type PartitionRaw struct {
	Type       string
	PTUUID     *string
	Name       string
	Rota       bool
	Size       uint64
	Model      *string
	UUID       *string
	PartUUID   *string
	FSType     *string
	FSSize     *uint64
	FSUsed     *uint64
	Mountpoint *string
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
	case "linux":
		return GetDisksLinux()
	default:
		return []Disk{}
	}
}

func GetDisksLinux() []Disk {
	disksStdout, err := exec.Command("lsblk", "-Jbo", "TYPE,PTUUID,NAME,ROTA,SIZE,MODEL,UUID,PARTUUID,FSTYPE,FSSIZE,FSUSED,MOUNTPOINT").Output()
	if err != nil {
		SleepyWarnLn("Failed to get disks! (%s)", err.Error())
		return []Disk{}
	}

	var disksRaw DisksRaw
	err = json.Unmarshal(disksStdout, &disksRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse disks! (%s)", err.Error())
		return []Disk{}
	}

	var disks []Disk
	disks = ArrayMap(disksRaw.Blockdevices, func(diskRaw DiskRaw) Disk {
		var disk Disk = Disk{
			Name:   diskRaw.Name,
			SSD:    !diskRaw.Rota,
			PTUUID: diskRaw.PTUUID,
			Size:   diskRaw.Size,
			Model:  diskRaw.Model,
		}
		if diskRaw.PTUUID == nil {
			disk.ID = GetMD5Hash(*diskRaw.Children[0].UUID)
		} else {
			disk.ID = GetMD5Hash(*diskRaw.PTUUID)
		}
		disk.Children = ArrayMap(diskRaw.Children, func(partRaw PartitionRaw) Partition {
			var part Partition = Partition{
				Name:       partRaw.Name,
				UUID:       partRaw.UUID,
				PartUUID:   partRaw.PartUUID,
				Type:       partRaw.FSType,
				Size:       partRaw.Size,
				Used:       partRaw.FSUsed,
				Mountpoint: partRaw.Mountpoint,
			}
			if partRaw.UUID == nil {
				part.ID = GetMD5Hash(disk.ID + *partRaw.PartUUID)
			} else {
				part.ID = GetMD5Hash(*partRaw.UUID)
			}
			if partRaw.FSSize != nil {
				part.Size = *partRaw.FSSize
			}

			return part
		})

		return disk
	})

	return disks
}
