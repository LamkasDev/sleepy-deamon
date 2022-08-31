package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/exp/slices"
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
	ID       string  `json:"id"`
	Parent   *string `json:"parent"`
	Image    string  `json:"image"`
	Creation int64   `json:"creation"`
	Ports    string  `json:"ports"`
	Status   string  `json:"status"`
	Names    string  `json:"names"`
	Mounts   string  `json:"mounts"`
	Networks string  `json:"networks"`
}

type ContainerDetailsRaw struct {
	StartedAt string `json:"State.StartedAt"`
}

type ContainerProjectRaw struct {
	Name        string
	Status      string
	ConfigFiles string
}

type ContainerProjectContainer struct {
	ID   string
	Name string
}

type ContainerProject struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Status     string   `json:"status"`
	Path       string   `json:"path"`
	Containers []string `json:"containers"`
}

func GetContainerProjects(handler *Handler) []ContainerProject {
	return GetContainerProjectsSystem(handler)
}

func GetContainerProjectsSystem(handler *Handler) []ContainerProject {
	if handler.Session == nil {
		SleepyWarnLn("Failed to get container projects! (%s)", "no session")
		return []ContainerProject{}
	}
	containerProjectsStdout, err := exec.Command("docker-compose", "ls", "--format", "json").Output()
	if err != nil {
		SleepyWarnLn("Failed to get container projects! (%s)", err.Error())
		return []ContainerProject{}
	}
	var containerProjectsRaw []ContainerProjectRaw
	err = json.Unmarshal([]byte(containerProjectsStdout), &containerProjectsRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse container projects! (%s)", err.Error())
		return []ContainerProject{}
	}

	var containerProjects []ContainerProject
	for _, containerProjectRaw := range containerProjectsRaw {
		containerProjectContainersCmd := exec.Command("docker-compose", "ps", "--format", "json")
		containerProjectContainersCmd.Dir = filepath.Dir(containerProjectRaw.ConfigFiles)
		containerProjectContainersStdout, err := containerProjectContainersCmd.Output()
		if err != nil {
			SleepyWarnLn("Failed to get container project containers! (%s)", err.Error())
			break
		}
		var containerProjectContainers []ContainerProjectContainer
		err = json.Unmarshal([]byte(containerProjectContainersStdout), &containerProjectContainers)
		if err != nil {
			SleepyWarnLn("Failed to parse container project containers! (%s)", err.Error())
			break
		}

		var containerProject ContainerProject = ContainerProject{
			ID:     GetMD5Hash(handler.Session.ID + containerProjectRaw.Name),
			Name:   containerProjectRaw.Name,
			Status: containerProjectRaw.Status,
			Path:   filepath.Dir(containerProjectRaw.ConfigFiles),
			Containers: ArrayMap(containerProjectContainers, func(containerProjectContainer ContainerProjectContainer) string {
				return containerProjectContainer.Name
			}),
		}

		containerProjects = append(containerProjects, containerProject)
	}

	return containerProjects
}

func GetContainers(containerProjects []ContainerProject) []Container {
	return GetContainersSystem(containerProjects)
}

func GetContainersSystem(containerProjects []ContainerProject) []Container {
	fields := []string{"ID", "Image", "Ports", "Status", "Names", "Mounts", "Networks"}
	fieldsCmd := strings.Join(ArrayMap(fields, func(e string) string { return fmt.Sprintf(`"%s":"{{.%s}}"`, e, e) }), ",")
	containersStdout, err := exec.Command("docker", "ps", "--format", fmt.Sprintf("{%s}", fieldsCmd)).Output()
	if err != nil {
		SleepyWarnLn("Failed to get containers! (%s)", err.Error())
		return []Container{}
	}
	containersStdoutMod := strings.ReplaceAll(string(containersStdout), "\n", ",")
	var containersRaw []ContainerRaw
	err = json.Unmarshal([]byte(fmt.Sprintf("[%s]", containersStdoutMod[:len(containersStdoutMod)-1])), &containersRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse containers! (%s)", err.Error())
		return []Container{}
	}

	// TODO: use coroutines
	var containers []Container
	for _, containerRaw := range containersRaw {
		detailedFields := []string{"State.StartedAt"}
		detailedFieldsCmd := strings.Join(ArrayMap(detailedFields, func(e string) string { return fmt.Sprintf(`"%s":"{{.%s}}"`, e, e) }), ",")
		containerStdout, err := exec.Command("docker", "inspect", "--format", fmt.Sprintf("{%s}", detailedFieldsCmd), containerRaw.ID).Output()
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

		var parent string
		for _, containerProject := range containerProjects {
			if slices.Contains(containerProject.Containers, containerRaw.Names) {
				parent = containerProject.ID
			}
		}

		var container Container = Container{
			ID:       GetMD5Hash(containerRaw.ID),
			Image:    containerRaw.Image,
			Creation: containerStartedAt.Unix(),
			Ports:    containerRaw.Ports,
			Status:   containerRaw.Status,
			Names:    containerRaw.Names,
			Mounts:   containerRaw.Mounts,
			Networks: containerRaw.Networks,
		}
		if parent == "" {
			container.Parent = nil
		} else {
			container.Parent = &parent
		}
		containers = append(containers, container)
	}

	return containers
}
