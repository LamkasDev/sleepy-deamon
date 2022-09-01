//go:build linux
// +build linux

package main

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

func GetNetworkUsageSystem() NetworkUsage {
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

		network.RX += int64(rx)
		network.TX += int64(tx)
	}

	return network
}
