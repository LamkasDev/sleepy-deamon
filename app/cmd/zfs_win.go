//go:build windows
// +build windows

package main

func GetZFSPoolsSystem(disks []Disk) []ZFSPool {
	return []ZFSPool{}
}

func SetZFSOptionSystem(name string, key string, value string) bool {
	return false
}

func GetZFSVersionSystem() *string {
	return nil
}
