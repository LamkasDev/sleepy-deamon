//go:build windows
// +build windows

package main

func GetZFSPoolsSystem(disks []Disk) []ZFSPool {
	return []ZFSPool{}
}
