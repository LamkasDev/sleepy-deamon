package main

import "runtime"

type Software struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func GetInstalledSoftware() []Software {
	software := []Software{}
	zfs := GetZFSVersion()
	if zfs != nil {
		software = append(software,
			Software{"zfs", *zfs},
		)
	}

	switch runtime.GOOS {
	case "linux", "windows":
		return GetInstalledSoftwareSystem()
	}

	return software
}
