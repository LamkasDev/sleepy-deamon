package main

import "runtime"

type Process struct {
	Name      string `json:"name"`
	Instances uint16 `json:"instances"`
	Memory    uint64 `json:"memory"`
}

func GetProcesses() []Process {
	switch runtime.GOOS {
	case "windows":
		return GetProcessesSystem()
	default:
		return []Process{}
	}
}
