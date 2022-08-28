//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"os/exec"
	"strconv"
)

type DiskWindowsRaw struct {
	DeviceId     string
	UniqueId     *string
	SerialNumber *string
	FriendlyName string
	Size         uint64
	Model        *string
	MediaType    string
	IsBoot       bool
	IsHidden     bool
}

type PartitionWindowsRaw struct {
	UniqueId    *string
	Guid        *string
	DiskNumber  int
	DriveLetter *string
	Size        uint64
}

func GetDisksSystem() []Disk {
	disksStdout, err := exec.Command("Powershell", "-Command", "Get-PhysicalDisk | ConvertTo-Json").Output()
	if err != nil {
		SleepyWarnLn("Failed to get disks! (%s)", err.Error())
		return []Disk{}
	}

	var disksRaw []DiskWindowsRaw
	err = json.Unmarshal(disksStdout, &disksRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse disks! (%s)", err.Error())
		return []Disk{}
	}

	partitionsStdout, err := exec.Command("Powershell", "-Command", "Get-Partition | ConvertTo-Json").Output()
	if err != nil {
		SleepyWarnLn("Failed to get partitions! (%s)", err.Error())
		return []Disk{}
	}

	var partitionsRaw []PartitionWindowsRaw
	err = json.Unmarshal(partitionsStdout, &partitionsRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse partitions! (%s)", err.Error())
		return []Disk{}
	}

	var disks []Disk
	for _, diskRaw := range disksRaw {
		if diskRaw.IsBoot || diskRaw.IsHidden {
			continue
		}

		var disk Disk = Disk{
			Name:   diskRaw.FriendlyName,
			SSD:    diskRaw.MediaType == "SSD",
			PTUUID: diskRaw.UniqueId,
			Size:   diskRaw.Size,
			Model:  diskRaw.Model,
		}
		if diskRaw.SerialNumber == nil {
			// Hope this works, windows's ids are shit
			disk.ID = GetMD5Hash(*diskRaw.UniqueId)
		} else {
			disk.ID = GetMD5Hash(*diskRaw.UniqueId + *diskRaw.SerialNumber)
		}
		disk.ID = GetMD5Hash(*diskRaw.UniqueId)
		for _, partRaw := range partitionsRaw {
			if strconv.Itoa(partRaw.DiskNumber) != diskRaw.DeviceId {
				continue
			}

			var part Partition = Partition{
				UUID:       partRaw.UniqueId,
				PartUUID:   nil,
				Type:       nil,
				Size:       partRaw.Size,
				Used:       nil,
				Mountpoint: partRaw.DriveLetter,
			}
			if partRaw.Guid == nil {
				// Hope this works, windows's ids are shit
				part.ID = GetMD5Hash(*partRaw.UniqueId)
			} else {
				part.ID = GetMD5Hash(*partRaw.UniqueId + *partRaw.Guid)
			}
			if partRaw.DriveLetter == nil {
				part.Name = "???"
			} else {
				part.Name = *partRaw.DriveLetter
			}

			disk.Children = append(disk.Children, part)
		}

		disks = append(disks, disk)
	}

	return disks
}
