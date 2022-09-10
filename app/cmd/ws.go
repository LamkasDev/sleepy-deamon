package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jwalton/gchalk"
)

const (
	WebsocketMessageTypeAuth        string = "DAEMON_AUTH"
	WebsocketMessageTypeAuthSuccess string = "DAEMON_AUTH_SUCCESS"
	WebsocketMessageTypeAuthFailure string = "DAEMON_AUTH_FAILURE"

	WebsocketMessageTypeRequestResources      string = "DAEMON_REQUEST_RESOURCES"
	WebsocketMessageTypeRequestResourcesReply string = "DAEMON_REQUEST_RESOURCES_REPLY"
	WebsocketMessageTypeRequestDatabaseBackup string = "DAEMON_REQUEST_DATABASE_BACKUP"

	WebsocketMessageTypeRequestStats      string = "DAEMON_REQUEST_STATS"
	WebsocketMessageTypeRequestStatsReply string = "DAEMON_REQUEST_STATS_REPLY"

	WebsocketMessageTypeTaskProgress string = "DAEMON_TASK_PROGRESS"
)

const (
	WebsocketAuthFailureWrongToken      string = "WRONG_TOKEN"
	WebsocketAuthFailureVersionMismatch string = "VERSION_MISMATCH"
)

type Session struct {
	ID   string
	Name string
}

type WebsocketMessage struct {
	Type string `json:"type"`
}

type WebsocketAuthMessage struct {
	Type      string   `json:"type"`
	Token     string   `json:"token"`
	Version   string   `json:"version"`
	Databases []string `json:"databases"`
}

type WebsocketAuthSuccessMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Name string `json:"name"`
}

type WebsocketAuthFailureMessage struct {
	Type   string `json:"type"`
	Reason string `json:"reason"`
}

type WebsocketAuthFailureVersionMismatchMessage struct {
	Type    string `json:"type"`
	Reason  string `json:"reason"`
	Version string `json:"version"`
}

type WebsocketRequestResourcesMessage struct {
	Type      string   `json:"type"`
	Resources []string `json:"resources"`
}

const (
	WebsocketResourcesGeneralType    string = "GENERAL"
	WebsocketResourcesContainersType string = "CONTAINERS"
	WebsocketResourcesDisksType      string = "DISKS"
)

type WebsocketRequestResourcesReplyMessage struct {
	Type              string                              `json:"type"`
	Memory            *MemoryState                        `json:"memory"`
	Software          []WebsocketRequestResourcesSoftware `json:"software"`
	Disks             []Disk                              `json:"disks"`
	ZFSPools          []ZFSPool                           `json:"zfsPools"`
	Containers        []Container                         `json:"containers"`
	ContainerProjects []ContainerProject                  `json:"containerProjects"`
}

type WebsocketRequestResourcesSoftware struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type WebsocketRequestDatabaseBackupMessage struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	Task     string `json:"task"`
	File     string `json:"file"`
}

type WebsocketRequestStatsReplyMessage struct {
	Type       string           `json:"type"`
	CPU        CPUUsage         `json:"cpu"`
	Memory     MemoryUsage      `json:"memory"`
	Disks      []DiskUsage      `json:"disks"`
	Network    NetworkUsage     `json:"network"`
	Containers []ContainerUsage `json:"containers"`
}

const (
	TaskStatusRunning  string = "RUNNING"
	TaskStatusFailed   string = "FAILED"
	TaskStatusFinished string = "FINISHED"
)

type WebsocketTaskProgressMessage struct {
	Type     string  `json:"type"`
	Task     string  `json:"task"`
	Progress float32 `json:"progress"`
	Status   string  `json:"status"`
}

func ConnectWebsocket(handler *Handler) *websocket.Conn {
	u := url.URL{Host: handler.Config.DaemonHost, Path: "/socket"}
	if handler.Config.Https {
		u.Scheme = "wss"
	} else {
		u.Scheme = "ws"
	}
	SleepyLogLn("Connecting to %s...", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		SleepyWarnLn("Failed to connect (%s)! Reconnecting in %d s...", err.Error(), handler.Config.ReconnectTimeout)
		return nil
	}
	SleepyLogLn("Connected!")

	authMessage := WebsocketAuthMessage{
		Type:      WebsocketMessageTypeAuth,
		Token:     handler.Config.Token,
		Version:   DaemonVersion,
		Databases: []string{},
	}
	for _, e := range handler.Config.DatabaseCredentials {
		for _, j := range e.Databases {
			authMessage.Databases = append(authMessage.Databases, j.ID)
		}
	}
	ws.WriteJSON(authMessage)

	return ws
}

func ProcessWebsocket(handler *Handler, ws *websocket.Conn) error {
	for {
		_, messageRaw, err := ws.ReadMessage()
		if err != nil {
			SleepyWarnLn("Disconnected (%s)! Reconnecting in %d s...", err.Error(), handler.Config.ReconnectTimeout)
			return err
		}
		var messageBase WebsocketMessage
		err = json.Unmarshal(messageRaw, &messageBase)
		if err != nil {
			SleepyWarnLn("Failed to parse websocket message! (%s)", err.Error())
			continue
		}
		SleepyLogLn("Got message of type %s", messageBase.Type)

		switch messageBase.Type {
		case WebsocketMessageTypeAuthSuccess:
			var message WebsocketAuthSuccessMessage
			_ = json.Unmarshal(messageRaw, &message)
			handler.Session = &Session{
				ID:   message.ID,
				Name: message.Name,
			}
			SleepyLogLn("Logged in as %s! (id: %s)", handler.Session.Name, handler.Session.ID)
		case WebsocketMessageTypeAuthFailure:
			var message WebsocketAuthFailureMessage
			_ = json.Unmarshal(messageRaw, &message)
			switch message.Reason {
			case WebsocketAuthFailureWrongToken:
				SleepyErrorLn("Incorrect token! Closing the deamon...")
				closeDaemon(handler)
			case WebsocketAuthFailureVersionMismatch:
				var message WebsocketAuthFailureVersionMismatchMessage
				_ = json.Unmarshal(messageRaw, &message)
				SleepyWarnLn("Version mismatch! Current version %s is not needed %s! Updating...", gchalk.Red(DaemonVersion), gchalk.Green(message.Version))
				err := Update(handler, message.Version)
				if err != nil {
					return err
				}
			default:
				return fmt.Errorf("failed to auth: %s", message.Reason)
			}
		case WebsocketMessageTypeRequestResources:
			var message WebsocketRequestResourcesMessage
			_ = json.Unmarshal(messageRaw, &message)

			requestResourcesReplyMessage := WebsocketRequestResourcesReplyMessage{
				Type: WebsocketMessageTypeRequestResourcesReply,
			}
			for _, resource := range message.Resources {
				switch resource {
				case WebsocketResourcesGeneralType:
					memory, _ := GetMemoryDetails()
					requestResourcesReplyMessage.Memory = &memory
					requestResourcesReplyMessage.Software = []WebsocketRequestResourcesSoftware{}
					zfs := GetZFSVersion()
					if zfs != nil {
						requestResourcesReplyMessage.Software = append(requestResourcesReplyMessage.Software,
							WebsocketRequestResourcesSoftware{"zfs", *zfs},
						)
					}
				case WebsocketResourcesContainersType:
					requestResourcesReplyMessage.Containers, requestResourcesReplyMessage.ContainerProjects = GetContainers(handler)
				case WebsocketResourcesDisksType:
					requestResourcesReplyMessage.Disks = GetDisks()
					requestResourcesReplyMessage.ZFSPools = GetZFSPools(requestResourcesReplyMessage.Disks)
				}
			}

			ws.WriteJSON(requestResourcesReplyMessage)
		case WebsocketMessageTypeRequestDatabaseBackup:
			var message WebsocketRequestDatabaseBackupMessage
			_ = json.Unmarshal(messageRaw, &message)

			taskProgressMessage := WebsocketTaskProgressMessage{
				Type:   WebsocketMessageTypeTaskProgress,
				Task:   message.Task,
				Status: TaskStatusRunning,
			}
			path, err := CreateBackup(handler, message.Database)
			if err != nil {
				SleepyWarnLn("Failed to create a database backup! (%s)", err.Error())
				taskProgressMessage.Status = TaskStatusFailed
				ws.WriteJSON(taskProgressMessage)
				continue
			}

			uploadFileData := UploadFileBackupDatabaseData{
				Type:     UploadFileDataBackupDatabase,
				Database: message.Database,
				Task:     message.Task,
			}
			err = UploadFile(handler, path, uploadFileData)
			if err != nil {
				SleepyWarnLn("Failed to upload database backup! (%s)", err.Error())
				taskProgressMessage.Status = TaskStatusFailed
				ws.WriteJSON(taskProgressMessage)
				continue
			}
			taskProgressMessage.Status = TaskStatusFinished
			taskProgressMessage.Progress = 100
			ws.WriteJSON(taskProgressMessage)
		case WebsocketMessageTypeRequestStats:
			requestStatsReplyMessage := GetStatsMessage(handler)
			requestStatsReplyMessage.Type = WebsocketMessageTypeRequestStatsReply
			ws.WriteJSON(requestStatsReplyMessage)
		}
	}
}

func GetStatsMessage(handler *Handler) WebsocketRequestStatsReplyMessage {
	timeDiff := uint64(time.Since(handler.StatsSnapshot.Timestamp).Seconds())
	handler.StatsSnapshot.Timestamp = time.Now()
	message := WebsocketRequestStatsReplyMessage{
		CPU:   CPUUsage{},
		Disks: []DiskUsage{},
	}

	_, memory := GetMemoryDetails()
	message.Memory = memory

	rawCpuUsage := GetCPUUsage()
	rawCpuTotal := float32(rawCpuUsage.Total - handler.StatsSnapshot.RawCPUUsage.Total)
	message.CPU = CPUUsage{
		User:   (float32(rawCpuUsage.User-handler.StatsSnapshot.RawCPUUsage.User) / rawCpuTotal) * 100,
		System: (float32(rawCpuUsage.System-handler.StatsSnapshot.RawCPUUsage.System) / rawCpuTotal) * 100,
	}
	handler.StatsSnapshot.RawCPUUsage = rawCpuUsage

	networkUsage := GetNetworkUsage()
	message.Network = NetworkUsage{
		RX: MathMin((networkUsage.RX-handler.StatsSnapshot.NetworkUsage.RX)/int64(timeDiff), 0),
		TX: MathMin((networkUsage.TX-handler.StatsSnapshot.NetworkUsage.TX)/int64(timeDiff), 0),
	}
	handler.StatsSnapshot.NetworkUsage = networkUsage

	containerUsagesSnapshot := GetContainerUsages()
	var containerUsages []ContainerUsage
	for _, containerUsageSnapshot := range containerUsagesSnapshot {
		lastContainerUsageIndex := -1
		for i, tempContainerUsage := range handler.StatsSnapshot.ContainerUsages {
			if containerUsageSnapshot.Parent == tempContainerUsage.Parent {
				lastContainerUsageIndex = i
			}
		}
		if lastContainerUsageIndex == -1 {
			continue
		}
		containerUsage := ContainerUsage{
			Parent: containerUsageSnapshot.Parent,
			RX:     MathMin((containerUsageSnapshot.RX-handler.StatsSnapshot.ContainerUsages[lastContainerUsageIndex].RX)/int64(timeDiff), 0),
			TX:     MathMin((containerUsageSnapshot.TX-handler.StatsSnapshot.ContainerUsages[lastContainerUsageIndex].TX)/int64(timeDiff), 0),
			CPU:    containerUsageSnapshot.CPU,
			Memory: containerUsageSnapshot.Memory,
			Read:   (containerUsageSnapshot.Read - handler.StatsSnapshot.ContainerUsages[lastContainerUsageIndex].Read) / uint64(timeDiff),
			Write:  (containerUsageSnapshot.Write - handler.StatsSnapshot.ContainerUsages[lastContainerUsageIndex].Write) / uint64(timeDiff),
		}
		containerUsages = append(containerUsages, containerUsage)
	}
	message.Containers = containerUsages
	handler.StatsSnapshot.ContainerUsages = containerUsagesSnapshot

	switch runtime.GOOS {
	case "linux":
		disks := GetDisks()
		diskUsagesSnapshot := GetDiskUsagesLinux()
		var diskUsages []DiskUsage
		for _, rawDiskUsage := range diskUsagesSnapshot {
			lastDiskUsageIndex := -1
			for i, diskUsage := range handler.StatsSnapshot.LinuxRawDiskUsages {
				if diskUsage.Name == rawDiskUsage.Name {
					lastDiskUsageIndex = i
				}
			}
			if lastDiskUsageIndex == -1 {
				continue
			}
			matchingDiskIndex := -1
			for i, disk := range disks {
				if disk.Name == rawDiskUsage.Name {
					matchingDiskIndex = i
				}
			}
			if matchingDiskIndex == -1 {
				continue
			}
			readsDiff := rawDiskUsage.Reads - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].Reads
			if readsDiff <= 0 {
				readsDiff = 1
			}
			writesDiff := rawDiskUsage.Writes - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].Writes
			if writesDiff <= 0 {
				writesDiff = 1
			}

			diskUsages = append(diskUsages, DiskUsage{
				Parent:       disks[matchingDiskIndex].ID,
				Read:         ((rawDiskUsage.ReadSectors - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].ReadSectors) * 512) / timeDiff,
				Write:        ((rawDiskUsage.WriteSectors - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].WriteSectors) * 512) / timeDiff,
				ReadLatency:  ((rawDiskUsage.ReadTime - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].ReadTime) / readsDiff),
				WriteLatency: ((rawDiskUsage.WriteTime - handler.StatsSnapshot.LinuxRawDiskUsages[lastDiskUsageIndex].WriteTime) / writesDiff),
			})
			// SleepyLogLn("Adding... (name: %s, id: %s, read: %v, write: %v, timeDiff: %v)", disks[matchingDiskIndex].Name, disks[matchingDiskIndex].ID, diskUsage.Read, diskUsage.Write, timeDiff)
		}
		message.Disks = diskUsages
		handler.StatsSnapshot.LinuxRawDiskUsages = diskUsagesSnapshot
	}

	return message
}
