//go:build windows
// +build windows

package main

import "syscall"

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	psapi    = syscall.NewLazyDLL("psapi.dll")
	shellapi = syscall.NewLazyDLL("shell32.dll")

	getSystemTimes       = kernel32.NewProc("GetSystemTimes")
	globalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")

	createFileW     = kernel32.NewProc("CreateFileW")
	deviceIoControl = kernel32.NewProc("DeviceIoControl")

	enumProcesses        = psapi.NewProc("EnumProcesses")
	enumProcessModules   = psapi.NewProc("EnumProcessModules")
	openProcess          = kernel32.NewProc("OpenProcess")
	getModuleBaseName    = psapi.NewProc("GetModuleBaseNameA")
	getModuleFileName    = psapi.NewProc("GetModuleFileNameExA")
	extractIcon          = shellapi.NewProc("ExtractAssociatedIconA")
	getProcessMemoryInfo = psapi.NewProc("GetProcessMemoryInfo")
	isProcessCritical    = kernel32.NewProc("IsProcessCritical")
)

var processQueryInformation = 0x0400
var processVmRead = 0x0010

type FileTimeWindows struct {
	Low  uint32
	High uint32
}
