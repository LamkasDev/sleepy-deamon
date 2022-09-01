//go:build linux
// +build linux

package main

import (
	"encoding/json"
	"os"
	"os/exec"
)

type DisksLinuxRaw struct {
	Blockdevices []DiskLinuxRaw
}

type DiskLinuxRaw struct {
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
	Children   []PartitionLinuxRaw
}

type PartitionLinuxRaw struct {
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

func GetDisksSystem() []Disk {
	if Geteuid() != 0 || Geteuid() != -1 {
		SleepyWarnLn("Skiping disks due to insufficient permissions!")
		return []Disk{}
	}
	disksStdout, err := exec.Command("lsblk", "-Jbo", "TYPE,PTUUID,NAME,ROTA,SIZE,MODEL,UUID,PARTUUID,FSTYPE,FSSIZE,FSUSED,MOUNTPOINT").Output()
	if err != nil {
		SleepyWarnLn("Failed to get disks! (%s)", err.Error())
		return []Disk{}
	}

	var disksRaw DisksLinuxRaw
	err = json.Unmarshal(disksStdout, &disksRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse disks! (%s)", err.Error())
		return []Disk{}
	}

	disks := ArrayMap(disksRaw.Blockdevices, func(diskRaw DiskLinuxRaw) Disk {
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
		disk.Children = ArrayMap(diskRaw.Children, func(partRaw PartitionLinuxRaw) Partition {
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
