//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"sync"
)

type DiskWindowsRaw struct {
	DeviceId     string
	UniqueId     *string
	SerialNumber *string
	FriendlyName string
	Size         uint64
	Model        *string
	MediaType    string
}

type PartitionWindowsRaw struct {
	UniqueId    *string
	Guid        *string
	DiskNumber  int
	DriveLetter *string
	Size        uint64
	IsBoot      bool
	IsHidden    bool
}

type VolumeWindowsRaw struct {
	DriveLetter     *string
	FileSystem      *string
	FileSystemLabel string
	SizeRemaining   uint64
}

func GetDisksSystem() []Disk {
	var wg sync.WaitGroup
	var disksRaw []DiskWindowsRaw
	var partitionsRaw []PartitionWindowsRaw
	var volumesRaw []VolumeWindowsRaw

	wg.Add(3)
	go func() {
		defer wg.Done()
		disksStdout, err := exec.Command("Powershell", "-Command", "Get-PhysicalDisk | ConvertTo-Json").Output()
		if err != nil {
			SleepyWarnLn("Failed to get disks! (%s)", err.Error())
			return
		}
		err = json.Unmarshal(disksStdout, &disksRaw)
		if err != nil {
			SleepyWarnLn("Failed to parse disks! (%s)", err.Error())
			return
		}
	}()
	go func() {
		defer wg.Done()
		partitionsStdout, err := exec.Command("Powershell", "-Command", "Get-Partition | ConvertTo-Json").Output()
		if err != nil {
			SleepyWarnLn("Failed to get partitions! (%s)", err.Error())
			return
		}
		err = json.Unmarshal(partitionsStdout, &partitionsRaw)
		if err != nil {
			SleepyWarnLn("Failed to parse partitions! (%s)", err.Error())
			return
		}
	}()
	go func() {
		defer wg.Done()
		volumesStdout, err := exec.Command("Powershell", "-Command", "Get-Volume | ConvertTo-Json").Output()
		if err != nil {
			SleepyWarnLn("Failed to get volumes! (%s)", err.Error())
			return
		}
		err = json.Unmarshal(volumesStdout, &volumesRaw)
		if err != nil {
			SleepyWarnLn("Failed to parse volumes! (%s)", err.Error())
			return
		}
	}()
	wg.Wait()

	var disks []Disk
	for _, diskRaw := range disksRaw {
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
		disk.Children = []Partition{}
		for _, partRaw := range partitionsRaw {
			if partRaw.IsHidden || strconv.Itoa(partRaw.DiskNumber) != diskRaw.DeviceId {
				continue
			}

			var part Partition = Partition{
				UUID:       partRaw.UniqueId,
				PartUUID:   nil,
				Type:       nil,
				Size:       partRaw.Size,
				Used:       nil,
				Mountpoint: partRaw.DriveLetter,
				Flags:      0,
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

				// Let's hope that 1 volume = 1 partition
				matchingVolumeIndex := -1
				for i, volumeRaw := range volumesRaw {
					if *partRaw.DriveLetter == *volumeRaw.DriveLetter {
						matchingVolumeIndex = i
					}
				}
				if matchingVolumeIndex != -1 {
					if volumesRaw[matchingVolumeIndex].FileSystemLabel != "" {
						part.Name = fmt.Sprintf("%s (%s)", volumesRaw[matchingVolumeIndex].FileSystemLabel, *partRaw.DriveLetter)
					}
					part.Type = volumesRaw[matchingVolumeIndex].FileSystem
					part.Used = &volumesRaw[matchingVolumeIndex].SizeRemaining
				}
			}
			if partRaw.IsBoot {
				part.Flags |= PartitionFlagBoot
			}

			disk.Children = append(disk.Children, part)
		}

		disks = append(disks, disk)
	}

	return disks
}
