package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type Container struct {
	ID       string `json:"id"`
	Image    string `json:"image"`
	Creation int64  `json:"creation"`
	Ports    string `json:"ports"`
	Status   string `json:"status"`
	Names    string `json:"names"`
	Mounts   string `json:"mounts"`
	Networks string `json:"networks"`
}

type ContainerDetailed struct {
	StartedAt string `json:"State.StartedAt"`
}

func GetContainers() []Container {
	return GetContainersAll()
}

// TODO: use coroutines
func GetContainersAll() []Container {
	fields := []string{"ID", "Image", "Ports", "Status", "Names", "Mounts", "Networks"}
	fieldsCmd := strings.Join(ArrayMap(fields, func(e string) string { return fmt.Sprintf(`"%s":"{{.%s}}"`, e, e) }), ",")
	containersStdout, err := exec.Command("docker", "ps", "--format", fmt.Sprintf("{%s}", fieldsCmd)).Output()
	if err != nil {
		SleepyWarnLn("Failed to get containers! (%s)", err.Error())
		return []Container{}
	}
	containersStdoutMod := strings.ReplaceAll(string(containersStdout), "\n", ",")
	var containersRaw []Container
	err = json.Unmarshal([]byte(fmt.Sprintf("[%s]", containersStdoutMod[:len(containersStdoutMod)-1])), &containersRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse containers! (%s)", err.Error())
		return []Container{}
	}

	for i := 0; i < len(containersRaw); i++ {
		container := &containersRaw[i]

		detailedFields := []string{"State.StartedAt"}
		detailedFieldsCmd := strings.Join(ArrayMap(detailedFields, func(e string) string { return fmt.Sprintf(`"%s":"{{.%s}}"`, e, e) }), ",")
		containerStdout, err := exec.Command("docker", "inspect", "--format", fmt.Sprintf("{%s}", detailedFieldsCmd), container.ID).Output()
		if err != nil {
			SleepyWarnLn("Failed to get container details! (%s)", err.Error())
			break
		}
		var containerDetailed ContainerDetailed
		err = json.Unmarshal([]byte(containerStdout), &containerDetailed)
		if err != nil {
			SleepyWarnLn("Failed to parse container details! (%s)", err.Error())
			break
		}
		containerStartedAt, err := time.Parse("2006-01-02T15:04:05.999999999", containerDetailed.StartedAt[:len(containerDetailed.StartedAt)-1])
		if err != nil {
			SleepyWarnLn("Failed to parse container start date! (%s)", err.Error())
			break
		}

		container.ID = GetMD5Hash(*&container.ID)
		container.Creation = containerStartedAt.Unix()
	}

	return containersRaw
}
