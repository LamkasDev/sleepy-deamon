package main

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
)

type DockerInfo struct {
	OperatingSystem string `json:"OperatingSystem"`
}

func GetDockerInfo(handler *Handler) DockerInfo {
	fields := []string{"OperatingSystem"}
	dockerStdout, err := exec.Command("docker", "info", "--format", GetDockerFormat(fields)).Output()
	if err != nil {
		SleepyWarnLn("Failed to get docker info! (%s)", err.Error())
		return DockerInfo{}
	}
	var dockerInfo DockerInfo
	err = json.Unmarshal(dockerStdout, &dockerInfo)
	if err != nil {
		SleepyWarnLn("Failed to parse docker info! (%s)", err.Error())
		return DockerInfo{}
	}

	return dockerInfo
}

func IsDockerDesktop(handler *Handler) bool {
	return handler.LastCache.DockerInfo.OperatingSystem == "Docker Desktop"
}

func ConvertDockerPath(handler *Handler, path string) string {
	if IsDockerDesktop(handler) {
		path = strings.ReplaceAll(path, "/var/lib/docker", "\\\\wsl.localhost\\docker-desktop-data\\data\\docker")
		path = filepath.FromSlash(path)
	}
	SleepyLogLn(path)

	return path
}
