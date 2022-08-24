package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type DiskUsage struct {
	Parent       string `json:"parent"`
	Read         uint64 `json:"read"`
	Write        uint64 `json:"write"`
	ReadLatency  uint64 `json:"readLatency"`
	WriteLatency uint64 `json:"writeLatency"`
}

type DiskUsageLinuxRaw struct {
	Name         string
	Reads        uint64
	ReadSectors  uint64
	ReadTime     uint64
	Writes       uint64
	WriteSectors uint64
	WriteTime    uint64
}

// TODO: make windows implementation

func GetDiskUsagesLinux() []DiskUsageLinuxRaw {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		SleepyWarnLn("Failed to get disks usage! (%s)", err.Error())
		return []DiskUsageLinuxRaw{}
	}
	defer file.Close()

	// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
	var disks []DiskUsageLinuxRaw
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 14 {
			continue
		}
		name := fields[2]
		reads, _ := strconv.ParseUint(fields[3], 10, 64)
		readSectors, _ := strconv.ParseUint(fields[5], 10, 64)
		readTime, _ := strconv.ParseUint(fields[6], 10, 64)
		writes, _ := strconv.ParseUint(fields[7], 10, 64)
		writeSectors, _ := strconv.ParseUint(fields[9], 10, 64)
		writeTime, _ := strconv.ParseUint(fields[10], 10, 64)

		disks = append(disks, DiskUsageLinuxRaw{
			Name:         name,
			Reads:        reads,
			ReadSectors:  readSectors,
			ReadTime:     readTime,
			Writes:       writes,
			WriteSectors: writeSectors,
			WriteTime:    writeTime,
		})
	}

	return disks
}
