package main

import (
	"bufio"
	"os"
	"runtime"
	"strconv"
	"strings"
)

type VnStatRaw struct {
	Interfaces []VnStatInterface
}

type VnStatInterface struct {
	Traffic VnStatInterfaceTraffic
}

type VnStatInterfaceTraffic struct {
	Total VnStatInterfaceTrafficTotal
}

type VnStatInterfaceTrafficTotal struct {
	RX uint64
	TX uint64
}

type NetworkUsage struct {
	RX uint64 `json:"rx"`
	TX uint64 `json:"tx"`
}

// TODO: make windows implementation

func GetNetworkUsage() NetworkUsage {
	switch runtime.GOOS {
	case "linux":
		return GetNetworkUsageLinux()
	default:
		return NetworkUsage{}
	}
}

// TODO: remake this parsing the /proc/net/netstat file

func GetNetworkUsageLinux() NetworkUsage {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		SleepyWarnLn("Failed to get network usage! (%s)", err.Error())
		return NetworkUsage{}
	}
	defer file.Close()

	var network NetworkUsage
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		i := strings.IndexRune(line, ':')
		if i < 0 {
			continue
		}
		name := strings.TrimSpace(line[:i])
		if name != "eth0" && name != "wlan0" {
			continue
		}

		fields := strings.Fields(line[i+1:])
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)

		network.RX += rx
		network.TX += tx
	}

	return network
}
