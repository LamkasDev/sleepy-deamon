//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type MemoryLinuxRaw struct {
	Total, Used, Buffers, Cached, Free, Available, Active, Inactive,
	SwapTotal, SwapUsed, SwapCached, SwapFree uint64
	MemAvailableEnabled bool
}

type MemoryStatLinux struct {
	Name string
	Ptr  *uint64
}

func GetMemorySystem() (MemoryLinuxRaw, error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		return MemoryLinuxRaw{}, err
	}
	defer file.Close()

	var memory MemoryLinuxRaw
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
		return MemoryLinuxRaw{}, err
	}

	memory.SwapUsed = memory.SwapTotal - memory.SwapFree
	if memory.MemAvailableEnabled {
		memory.Used = memory.Total - memory.Available
	} else {
		memory.Used = memory.Total - memory.Free - memory.Buffers - memory.Cached
	}

	return memory, nil
}

func GetMemoryDetailsSystem() (MemoryState, MemoryUsage) {
	memory, err := GetMemorySystem()
	if err != nil {
		SleepyWarnLn("Failed to get memory details! (%s)", err.Error())
		return MemoryState{}, MemoryUsage{}
	}
	return MemoryState{
			Total:     memory.Total,
			SwapTotal: memory.SwapTotal,
		}, MemoryUsage{
			Used:     (float32(memory.Used) / float32(memory.Total+1)) * 100,
			SwapUsed: (float32(memory.SwapUsed) / float32(memory.SwapTotal+1)) * 100,
		}
}
