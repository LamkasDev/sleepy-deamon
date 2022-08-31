//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strings"
	"strconv"
)

type MemoryUsageLinuxRaw struct {
	Total, Used, Buffers, Cached, Free, Available, Active, Inactive,
	SwapTotal, SwapUsed, SwapCached, SwapFree uint64
	MemAvailableEnabled bool
}

type MemoryStatLinux struct {
	Name string
	Ptr  *uint64
}

func GetMemoryUsageSystem() MemoryUsage {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		SleepyWarnLn("Failed to get RAM usage! (%s)", err.Error())
		return MemoryUsage{}
	}
	defer file.Close()

	var memory MemoryUsageLinuxRaw
	scanner := bufio.NewScanner(file)
	memStats := []MemoryStatLinux{
		{"MemTotal", &memory.Total},
		{"MemFree", &memory.Free},
		{"MemAvailable", &memory.Available},
		{"Buffers", &memory.Buffers},
		{"Cached", &memory.Cached},
		{"Active", &memory.Active},
		{"Inactive", &memory.Inactive},
		{"SwapCached", &memory.SwapCached},
		{"SwapTotal", &memory.SwapTotal},
		{"SwapFree", &memory.SwapFree},
	}
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.IndexRune(line, ':')
		if i < 0 {
			continue
		}
		fld := line[:i]
		for j, stat := range memStats {
			if stat.Name == fld {
				val := strings.TrimSpace(strings.TrimRight(line[i+1:], "kB"))
				if v, err := strconv.ParseUint(val, 10, 64); err == nil {
					*memStats[j].Ptr = v * 1024
				}
			}
		}
		if fld == "MemAvailable" {
			memory.MemAvailableEnabled = true
		}
	}
	if err := scanner.Err(); err != nil {
		SleepyWarnLn("Failed to scan /proc/meminfo!")
		return MemoryUsage{}
	}

	memory.SwapUsed = memory.SwapTotal - memory.SwapFree
	if memory.MemAvailableEnabled {
		memory.Used = memory.Total - memory.Available
	} else {
		memory.Used = memory.Total - memory.Free - memory.Buffers - memory.Cached
	}

	return MemoryUsage{
		Total:     memory.Total,
		Used:      memory.Used,
		SwapTotal: memory.SwapTotal,
		SwapUsed:  memory.SwapUsed,
	}
}
