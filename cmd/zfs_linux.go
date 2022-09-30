//go:build linux
// +build linux

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func GetZFSPoolsSystem(disks []Disk) []ZFSPool {
	zpoolStdout, err := exec.Command("zpool", "list", "-v", "-H", "-P").Output()
	if err != nil {
		SleepyWarnLn("Failed to get ZFS pools! (%s)", err.Error())
		return []ZFSPool{}
	}

	pools := []ZFSPool{}
	scanner := bufio.NewScanner(bytes.NewReader(zpoolStdout))
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		name := fields[0]
		size := ConvertToBytesShort(fields[1])
		used := ConvertToBytesShort(fields[2])

		if strings.HasPrefix(line, "\t") {
			lastNamePart := name[strings.LastIndex(name, "/")+1:]
			var parent string
			for _, disk := range disks {
				for _, partition := range disk.Children {
					if partition.Name == lastNamePart {
						parent = partition.ID
					}
				}
			}
			if parent == "" {
				SleepyWarnLn("Failed to find ZFS pool partition in disks! (name: %s)", lastNamePart)
				continue
			}

			partition := ZFSPartition{
				ID:   parent,
				Size: size,
				Used: used,
			}
			pools[len(pools)-1].Children = append(pools[len(pools)-1].Children, partition)
		} else {
			pool := ZFSPool{
				Name:     name,
				Size:     size,
				Used:     used,
				Children: []ZFSPartition{},
			}

			zfsDetailsStdout, err := exec.Command("zfs", "get", "all", name).Output()
			if err != nil {
				SleepyWarnLn("Failed to get ZFS pool details! (%s)", err.Error())
				return []ZFSPool{}
			}
			detailsScanner := bufio.NewScanner(bytes.NewReader(zfsDetailsStdout))
			detailsScanner.Scan()
			for detailsScanner.Scan() {
				detailsLine := detailsScanner.Text()
				detailsFields := strings.Fields(detailsLine)
				switch detailsFields[1] {
				case "guid":
					pool.ID = GetMD5Hash(detailsFields[2])
				case "compression":
					pool.Compression = &detailsFields[2]
				case "compressratio":
					n, _ := strconv.ParseFloat(detailsFields[2][:len(detailsFields[2])-1], 32)
					pool.CompressRatio = float32(n)
				case "encryption":
					pool.Encryption = detailsFields[2] == "on"
				case "atime":
					pool.ATime = detailsFields[2] == "on"
				case "version":
					n, _ := strconv.ParseInt(detailsFields[2], 10, 16)
					pool.Version = uint16(n)
				case "dedup":
					pool.Deduplication = detailsFields[2] == "on"
				case "relatime":
					pool.RelATime = detailsFields[2] == "on"
				}
			}

			pools = append(pools, pool)
		}
	}

	return pools
}

func SetZFSOptionSystem(name string, key string, value string) bool {
	_, err := exec.Command("zfs", "set", fmt.Sprintf("%s=%s", key, value), name).Output()
	if err != nil {
		return false
	}

	return true
}

func GetZFSVersionSystem() *string {
	zfsStdout, err := exec.Command("zfs", "--version").Output()
	if err != nil {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(zfsStdout))
	scanner.Scan()
	version := scanner.Text()
	return &version
}
