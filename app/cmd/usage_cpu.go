package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"unicode"
)

type CPUUsage struct {
	Total float32 `json:"total"`
}

type CPUUsageLinuxRaw struct {
	User, Nice, System, Idle, Iowait, Irq, Softirq, Steal, Guest, GuestNice, Total uint64
	CPUCount, StatCount                                                            int
}

type CPUStatLinux struct {
	Name string
	Ptr  *uint64
}

// TODO: make windows implementation

func GetCPUUsageLinux() CPUUsageLinuxRaw {
	file, err := os.Open("/proc/stat")
	if err != nil {
		SleepyWarnLn("Failed to get CPU usage! (%s)", err.Error())
		return CPUUsageLinuxRaw{}
	}
	defer file.Close()

	var cpu CPUUsageLinuxRaw
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
		return CPUUsageLinuxRaw{}
	}

	scanFields := strings.Fields(scanner.Text())[1:]
	cpu.StatCount = len(scanFields)
	for i, field := range scanFields {
		value, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			SleepyWarnLn("Failed to scan %s from /proc/stat!", cpuStats[i].Name)
			return CPUUsageLinuxRaw{}
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
