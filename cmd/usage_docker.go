package main

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type ContainerUsageRaw struct {
	ID       string
	CPUPerc  string
	MemUsage string
	NetIO    string
	BlockIO  string
}

type ContainerUsage struct {
	Parent string  `json:"parent"`
	RX     int64   `json:"rx"`
	TX     int64   `json:"tx"`
	CPU    float32 `json:"cpu"`
	Memory uint64  `json:"memory"`
	Read   uint64  `json:"read"`
	Write  uint64  `json:"write"`
}

func GetContainerUsages() []ContainerUsage {
	return GetContainerUsagesSystem()
}

func GetContainerUsagesSystem() []ContainerUsage {
	fields := []string{"ID", "CPUPerc", "MemUsage", "NetIO", "BlockIO"}
	containerUsagesStdout, err := exec.Command("docker", "stats", "--no-stream", "--format", GetDockerFormat(fields)).Output()
	if err != nil {
		SleepyWarnLn("Failed to get container usages! (%s)", err.Error())
		return []ContainerUsage{}
	}
	containerUsagesStdoutMod := strings.ReplaceAll(string(containerUsagesStdout), "\n", ",")
	var containerUsagesRaw []ContainerUsageRaw
	err = json.Unmarshal([]byte(fmt.Sprintf("[%s]", containerUsagesStdoutMod[:len(containerUsagesStdoutMod)-1])), &containerUsagesRaw)
	if err != nil {
		SleepyWarnLn("Failed to parse container usages! (%s)", err.Error())
		return []ContainerUsage{}
	}

	containerUsages := ArrayMap(containerUsagesRaw, func(containerUsageRaw ContainerUsageRaw) ContainerUsage {
		cpu, _ := strconv.ParseFloat(containerUsageRaw.CPUPerc[:len(containerUsageRaw.CPUPerc)-1], 32)
		memUsedRaw := containerUsageRaw.MemUsage[:strings.Index(containerUsageRaw.MemUsage, "/")]
		memUsed := ConvertToBytes(strings.Trim(memUsedRaw, " "))
		rxRaw := containerUsageRaw.NetIO[:strings.Index(containerUsageRaw.NetIO, "/")]
		rx := ConvertToBytes(strings.Trim(rxRaw, " "))
		txRaw := containerUsageRaw.NetIO[strings.Index(containerUsageRaw.NetIO, "/")+1:]
		tx := ConvertToBytes(strings.Trim(txRaw, " "))
		readRaw := containerUsageRaw.BlockIO[:strings.Index(containerUsageRaw.BlockIO, "/")]
		read := ConvertToBytes(strings.Trim(readRaw, " "))
		writeRaw := containerUsageRaw.BlockIO[strings.Index(containerUsageRaw.BlockIO, "/")+1:]
		write := ConvertToBytes(strings.Trim(writeRaw, " "))
		return ContainerUsage{
			Parent: GetMD5Hash(containerUsageRaw.ID),
			RX:     int64(rx),
			TX:     int64(tx),
			CPU:    float32(cpu),
			Memory: memUsed,
			Read:   read,
			Write:  write,
		}
	})
	return containerUsages
}
