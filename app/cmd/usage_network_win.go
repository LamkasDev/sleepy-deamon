//go:build windows
// +build windows

package main

import (
	"encoding/json"
	"os/exec"
)

type NetworkAdapterWindowsRaw struct {
	InterfaceAlias string
	ReceivedBytes  uint64
	SentBytes      uint64
}

func GetNetworkUsageSystem() NetworkUsage {
	networkAdaptersStdout, err := exec.Command("Powershell", "-Command", "Get-NetAdapterStatistics | ConvertTo-Json").Output()
	if err != nil {
		SleepyWarnLn("Failed to get network adapters! (%s)", err.Error())
		return NetworkUsage{}
	}

	var networkAdaptersRaw []NetworkAdapterWindowsRaw
	err = json.Unmarshal(networkAdaptersStdout, &networkAdaptersRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse network adapters! (%s)", err.Error())
		return NetworkUsage{}
	}

	var network NetworkUsage
	for _, adapter := range networkAdaptersRaw {
		if adapter.InterfaceAlias != "Ethernet" {
			continue
		}

		network.RX += adapter.ReceivedBytes
		network.TX += adapter.SentBytes
	}

	return network
}
