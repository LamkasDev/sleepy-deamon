package main

import "runtime"

type Process struct {
	Name   string
	Memory MemoryUsage
}

func GetProcessList() []Process {
	switch runtime.GOOS {
	case "windows":
		return GetProcessListSystem()
	default:
		return []Process{}
	}
}
