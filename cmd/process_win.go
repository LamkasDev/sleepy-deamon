//go:build windows
// +build windows

package main

import (
	"unsafe"

	"golang.org/x/exp/maps"
)

type ProcessMemoryUsageWindowsRaw struct {
	Size                       uint32
	PageFaultCount             uint32
	PeakWorkingSetSize         uint64
	WorkingSetSize             uint64
	QuotaPeakPagedPoolUsage    uint64
	QuotaPagedPoolUsage        uint64
	QuotaPeakNonPagedPoolUsage uint64
	QuotaNonPagedPoolUsage     uint64
	PagefileUsage              uint64
	PeakPagefileUsage          uint64
}

func GetProcessListSystem() []Process {
	var processListRaw [512]uint64
	var listBytes uint64
	ret, _, err := enumProcesses.Call(uintptr(unsafe.Pointer(&processListRaw)), unsafe.Sizeof(processListRaw), uintptr(unsafe.Pointer(&listBytes)))
	if ret == 0 {
		SleepyErrorLn("Failed to get process list! (%s)", err.Error())
		return []Process{}
	}

	processList := make(map[string]Process)
	for _, pid := range processListRaw {
		// Get process handle
		if pid == 0 {
			continue
		}
		prHandle, _, _ := openProcess.Call(uintptr(processQueryInformation+processVmRead), 0, uintptr(pid))
		if prHandle == 0 {
			// SleepyErrorLn("Failed to open process! (%s)", err.Error())
			continue
		}

		// Filter out critical processes
		var critical uint32
		ret, _, _ := isProcessCritical.Call(prHandle, uintptr(unsafe.Pointer(&critical)))
		if ret == 0 || critical == 1 {
			continue
		}

		// Get process name
		var modules [1]uintptr
		var modulesBytes uint64
		ret, _, err = enumProcessModules.Call(prHandle, uintptr(unsafe.Pointer(&modules)), unsafe.Sizeof(modules), uintptr(unsafe.Pointer(&modulesBytes)))
		if ret == 0 {
			SleepyErrorLn("Failed to enum process modules! (%s)", err.Error())
			continue
		}
		var modNameRaw [256]byte
		ret, _, err = getModuleBaseName.Call(prHandle, modules[0], uintptr(unsafe.Pointer(&modNameRaw)), 256)
		if ret == 0 {
			SleepyErrorLn("Failed to get base module name! (%s)", err.Error())
			continue
		}
		modName := string(modNameRaw[:])

		// Get memory usage
		var memory ProcessMemoryUsageWindowsRaw
		ret, _, err = getProcessMemoryInfo.Call(prHandle, uintptr(unsafe.Pointer(&memory)), unsafe.Sizeof(memory))
		if ret == 0 {
			SleepyErrorLn("Failed to get memory usage! (%s)", err.Error())
			continue
		}

		process, ok := processList[modName]
		if !ok {
			process = Process{
				Name: modName,
				Memory: MemoryUsage{
					Used:     0,
					SwapUsed: 0,
				},
			}
		}
		process.Memory.Used += float32(memory.WorkingSetSize - memory.PagefileUsage)
		process.Memory.SwapUsed += float32(memory.PagefileUsage)
		processList[modName] = process
	}
	return maps.Values(processList)
}
