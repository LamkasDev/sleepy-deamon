package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"golang.org/x/exp/maps"
)

type ContainerRaw struct {
	ID       string
	Image    string
	Creation int64
	Ports    string
	Status   string
	Names    string
	Mounts   string
	Networks string
}

type Container struct {
	ID        string  `json:"id"`
	Parent    *string `json:"parent"`
	Image     string  `json:"image"`
	Creation  int64   `json:"creation"`
	Ports     string  `json:"ports"`
	Status    string  `json:"status"`
	Names     string  `json:"names"`
	Mounts    string  `json:"mounts"`
	Networks  string  `json:"networks"`
	Directory string  `json:"directory"`
}

type ContainerDetailsRaw struct {
	StartedAt string                    `json:"State.StartedAt"`
	Status    string                    `json:"State.Status"`
	Labels    ContainerDetailsLabelsRaw `json:"Config.Labels"`
}

type ContainerDetailsLabelsRaw struct {
	ConfigHash *string `json:"com.docker.compose.config-hash"`
	Directory  *string `json:"com.docker.compose.project.working_dir"`
	Service    *string `json:"com.docker.compose.project"`
}

type ContainerProject struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Path   string `json:"path"`
}

func GetContainers(handler *Handler) ([]Container, []ContainerProject) {
	return GetContainersSystem(handler)
}

func GetContainersSystem(handler *Handler) ([]Container, []ContainerProject) {
	if handler.Session == nil {
		SleepyWarnLn("Failed to get containers! (%s)", "no session")
		return []Container{}, []ContainerProject{}
	}
	fields := []string{"ID", "Image", "Ports", "Status", "Names", "Mounts", "Networks"}
	containersStdout, err := exec.Command("docker", "ps", "--format", GetDockerFormat(fields)).Output()
	if err != nil {
		SleepyWarnLn("Failed to get containers! (%s)", err.Error())
		return []Container{}, []ContainerProject{}
	}
	containersStdoutMod := strings.ReplaceAll(string(containersStdout), "\n", ",")
	var containersRaw []ContainerRaw
	err = json.Unmarshal([]byte(fmt.Sprintf("[%s]", containersStdoutMod[:len(containersStdoutMod)-1])), &containersRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse containers! (%s)", err.Error())
		return []Container{}, []ContainerProject{}
	}

	// TODO: use coroutines
	containerProjects := make(map[string]ContainerProject)
	containers := []Container{}
	for _, containerRaw := range containersRaw {
		detailedFields := `{"State.StartedAt":"{{.State.StartedAt}}","State.Status":"{{.State.Status}}","Config.Labels":{{json .Config.Labels}}}`
		containerStdout, err := exec.Command("docker", "inspect", "--format", detailedFields, containerRaw.ID).Output()
		if err != nil {
			SleepyWarnLn("Failed to get container details! (%s)", err.Error())
			break
		}
		var containerDetailed ContainerDetailsRaw
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

		var container Container = Container{
			ID:       GetMD5Hash(containerRaw.ID),
			Image:    containerRaw.Image,
			Creation: containerStartedAt.Unix(),
			Ports:    containerRaw.Ports,
			Status:   containerDetailed.Status,
			Names:    containerRaw.Names,
			Mounts:   containerRaw.Mounts,
			Networks: containerRaw.Networks,
		}
		if containerDetailed.Labels.Directory != nil {
			id := GetMD5Hash(handler.Session.ID + *containerDetailed.Labels.Service)
			containerProject, ok := containerProjects[id]
			if !ok {
				containerProject = ContainerProject{
					ID:     id,
					Name:   *containerDetailed.Labels.Service,
					Status: "running",
					Path:   *containerDetailed.Labels.Directory,
				}
			}
			containerProjects[id] = containerProject
			container.Parent = &id
		}
		containers = append(containers, container)
	}

	return containers, maps.Values(containerProjects)
}
