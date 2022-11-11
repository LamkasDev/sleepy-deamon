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

func GetProcessesSystem() []Process {
	/* handle, err := syscall.GetCurrentProcess()
	if err != nil {
		SleepyErrorLn("Failed to get handle to current process! (%s)", err.Error())
		return []Process{}
	} */

	var processesRaw [512]uint64
	var listBytes uint64
	ret, _, err := enumProcesses.Call(uintptr(unsafe.Pointer(&processesRaw)), unsafe.Sizeof(processesRaw), uintptr(unsafe.Pointer(&listBytes)))
	if ret == 0 {
		SleepyErrorLn("Failed to get process list! (%s)", err.Error())
		return []Process{}
	}

	processes := make(map[string]Process)
	for _, pid := range processesRaw {
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
		/* var modPathRaw [2048]byte
		ret, _, err = getModuleFileName.Call(prHandle, modules[0], uintptr(unsafe.Pointer(&modPathRaw)), 2048)
		if ret == 0 {
			SleepyErrorLn("Failed to get base module path! (%s)", err.Error())
			continue
		}
		modPath := string(modPathRaw[:]) */

		// Get icon
		/* iconPathRaw := modPathRaw
		iconIndex := uint16(0)
		ret, _, err = extractIcon.Call(uintptr(handle), uintptr(unsafe.Pointer(&iconPathRaw)), uintptr(unsafe.Pointer(&iconIndex)))
		if ret == 0 {
			SleepyErrorLn("Failed to get process icon! (%s)", err.Error())
			continue
		}
		iconPath := string(iconPathRaw[:]) */

		// Get memory usage
		var memory ProcessMemoryUsageWindowsRaw
		ret, _, err = getProcessMemoryInfo.Call(prHandle, uintptr(unsafe.Pointer(&memory)), unsafe.Sizeof(memory))
		if ret == 0 {
			SleepyErrorLn("Failed to get memory usage! (%s)", err.Error())
			continue
		}

		process, ok := processes[modName]
		if !ok {
			process = Process{
				Name:      modName,
				Instances: 0,
				Memory:    0,
			}
		}
		process.Instances++
		process.Memory += memory.WorkingSetSize + memory.PagefileUsage
		processes[modName] = process
	}
	return maps.Values(processes)
}
