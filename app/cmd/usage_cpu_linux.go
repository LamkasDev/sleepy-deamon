//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type CPUStatLinux struct {
	Name string
	Ptr  *uint64
}

func GetCPUUsageSystem() CPUUsageRaw {
	file, err := os.Open("/proc/stat")
	if err != nil {
		SleepyWarnLn("Failed to get CPU usage! (%s)", err.Error())
		return CPUUsageRaw{}
	}
	defer file.Close()

	var cpu CPUUsageRaw
	scanner := bufio.NewScanner(file)
	cpuStats := []CPUStatLinux{
		{"user", &cpu.User},
		{"nice", &cpu.Nice},
		{"system", &cpu.System},
		{"idle", &cpu.Idle},
		{"iowait", &cpu.Iowait},
		{"irq", &cpu.Irq},
		{"softirq", &cpu.Softirq},
		{"steal", &cpu.Steal},
		{"guest", &cpu.Guest},
		{"guest_nice", &cpu.GuestNice},
	}
	if !scanner.Scan() {
		SleepyWarnLn("Failed to scan /proc/stat!")
		return CPUUsageRaw{}
	}

	scanFields := strings.Fields(scanner.Text())[1:]
	cpu.StatCount = len(scanFields)
	for i, field := range scanFields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			SleepyWarnLn("Failed to scan %s from /proc/stat!", cpuStats[i].Name)
			return CPUUsageRaw{}
		}
		*cpuStats[i].Ptr = value
		cpu.Total += value
	}
	// included in cpustat[CPUTIME_USER]
	cpu.Total -= cpu.Guest
	// included in cpustat[CPUTIME_NICE]
	cpu.Total -= cpu.GuestNice

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "cpu") && unicode.IsDigit(rune(line[3])) {
			cpu.CPUCount++
		}
	}
	if err := scanner.Err(); err != nil {
		SleepyWarnLn("Failed to scan /proc/stat! (%s)", err.Error())
	}

	return cpu
}
