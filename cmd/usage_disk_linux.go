//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func GetDiskUsagesSystem() []DiskUsageRaw {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		SleepyWarnLn("Failed to get disks usage! (%s)", err.Error())
		return []DiskUsageRaw{}
	}
	defer file.Close()

	// https://www.kernel.org/doc/Documentation/ABI/testing/procfs-diskstats
	var disks []DiskUsageRaw
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

		disks = append(disks, DiskUsageRaw{
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
