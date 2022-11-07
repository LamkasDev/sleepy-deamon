package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
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

	WebsocketMessageTypeConnectContainerLog    string = "DAEMON_CONNECT_CONTAINER_LOG"
	WebsocketMessageTypeRequestContainerLog    string = "DAEMON_REQUEST_CONTAINER_LOG"
	WebsocketMessageTypeDisconnectContainerLog string = "DAEMON_DISCONNECT_CONTAINER_LOG"
	WebsocketMessageTypeContainerLogMessage    string = "DAEMON_CONTAINER_LOG_MESSAGE"

	WebsocketMessageTypeBuildSmbConfig string = "DAEMON_BUILD_SMB_CONFIG"
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
	ZFS               []ZFSPool                           `json:"zfs"`
	Containers        []Container                         `json:"containers"`
	ContainerProjects []ContainerProject                  `json:"containerProjects"`
	ProcessList       []Process                           `json:"processList"`
}

type WebsocketRequestResourcesSoftware struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type WebsocketRequestDatabaseBackupMessage struct {
	Type     string `json:"type"`
	Database string `json:"database"`
	Data     bool   `json:"data"`
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
	ID       string  `json:"id"`
	Progress float32 `json:"progress"`
	Status   string  `json:"status"`
}

type WebsocketConnectContainerLogMessage struct {
	Type      string                             `json:"type"`
	Container WebsocketConnectContainerContainer `json:"container"`
	Options   WebsocketConnectContainerOptions   `json:"options"`
}
type WebsocketConnectContainerContainer struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Path *string `json:"path"`
}
type WebsocketConnectContainerOptions struct {
	Project bool   `json:"project"`
	Tail    uint32 `json:"tail"`
}

type WebsocketRequestContainerLogMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
	Task string `json:"task"`
}

type WebsocketDisconnectContainerLogMessage struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type WebsocketContainerLogMessageMessage struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Message string `json:"message"`
}

type WebsocketBuildSmbConfigMessage struct {
	Type   string `json:"type"`
	Config string `json:"config"`
}

func ConnectWebsocket(handler *Handler) *websocket.Conn {
	u := url.URL{Host: handler.Config.DaemonHost, Path: "/socket", Scheme: "wss"}
	SleepyLogLn("Connecting to %s...", u.String())

	ws, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		SleepyWarnLn("Failed to connect (%s)! Reconnecting in %d s...", err.Error(), handler.Config.ReconnectTimeout)
		return nil
	}
	SleepyLogLn("Connected!")
	return ws
}

func AuthWebsocket(handler *Handler) {
	authMessage := WebsocketAuthMessage{
		Type:      WebsocketMessageTypeAuth,
		Token:     handler.Config.Token,
		Version:   DaemonVersion,
		Databases: []string{},
	}
	for _, e := range handler.Credentials.Databases {
		for _, j := range e.Databases {
			authMessage.Databases = append(authMessage.Databases, j.ID)
		}
	}
	SendWebsocketMessage(handler, authMessage)
}

func SendWebsocketMessage(handler *Handler, message any) error {
	handler.WSMutex.Lock()
	defer handler.WSMutex.Unlock()
	return handler.WS.WriteJSON(message)
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
			InitSnapshot(handler)
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
			requestResourcesReplyMessage := GetResourcesMessage(handler, message.Resources)
			SendWebsocketMessage(handler, requestResourcesReplyMessage)
		case WebsocketMessageTypeRequestDatabaseBackup:
			var message WebsocketRequestDatabaseBackupMessage
			_ = json.Unmarshal(messageRaw, &message)

			taskProgressMessage := WebsocketTaskProgressMessage{
				Type:   WebsocketMessageTypeTaskProgress,
				ID:     message.Task,
				Status: TaskStatusRunning,
			}
			var path string
			if message.Data {
				path, err = CreateBackup(handler, message.Database)
			} else {
				path, err = CreateBackup(handler, message.Database, "--no-data")
			}
			if err != nil {
				SleepyWarnLn("Failed to create a database backup! (%s)", err.Error())
				taskProgressMessage.Status = TaskStatusFailed
				SendWebsocketMessage(handler, taskProgressMessage)
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
				SendWebsocketMessage(handler, taskProgressMessage)
				continue
			}
			taskProgressMessage.Status = TaskStatusFinished
			taskProgressMessage.Progress = 100
			SendWebsocketMessage(handler, taskProgressMessage)
		case WebsocketMessageTypeRequestStats:
			requestStatsReplyMessage := GetStatsMessage(handler)
			requestStatsReplyMessage.Type = WebsocketMessageTypeRequestStatsReply
			SendWebsocketMessage(handler, requestStatsReplyMessage)
		case WebsocketMessageTypeConnectContainerLog:
			var message WebsocketConnectContainerLogMessage
			_ = json.Unmarshal(messageRaw, &message)

			ConnectContainerLogger(handler, message.Container, message.Options)
		case WebsocketMessageTypeRequestContainerLog:
			var message WebsocketRequestContainerLogMessage
			_ = json.Unmarshal(messageRaw, &message)

			for _, container := range handler.LastCache.Containers {
				if container.ID == message.ID {
					RequestContainerLog(handler, container, message.Task)
				}
			}
		case WebsocketMessageTypeDisconnectContainerLog:
			var message WebsocketDisconnectContainerLogMessage
			_ = json.Unmarshal(messageRaw, &message)

			DisconnectContainerLogger(handler, message.ID)
		case WebsocketMessageTypeBuildSmbConfig:
			var message WebsocketBuildSmbConfigMessage
			_ = json.Unmarshal(messageRaw, &message)

			RebuildSmbConfig(handler, message.Config)
		}
	}
}

func GetResourcesMessage(handler *Handler, resources []string) WebsocketRequestResourcesReplyMessage {
	message := WebsocketRequestResourcesReplyMessage{
		Type: WebsocketMessageTypeRequestResourcesReply,
	}

	var wg sync.WaitGroup
	for _, resource := range resources {
		wg.Add(1)
		go func(resource string) {
			defer wg.Done()
			switch resource {
			case WebsocketResourcesGeneralType:
				memory, _ := GetMemoryDetails()
				message.Memory = &memory
				message.Software = []WebsocketRequestResourcesSoftware{}
				zfs := GetZFSVersion()
				if zfs != nil {
					message.Software = append(message.Software,
						WebsocketRequestResourcesSoftware{"zfs", *zfs},
					)
				}
			case WebsocketResourcesContainersType:
				message.Containers, message.ContainerProjects = GetContainers(handler)
				handler.LastCache.Containers = message.Containers
				handler.LastCache.ContainerProjects = message.ContainerProjects
			case WebsocketResourcesDisksType:
				message.Disks = GetDisks()
				message.ZFS = GetZFSPools(message.Disks)
			}
		}(resource)
	}
	wg.Wait()

	return message
}

func GetStatsMessage(handler *Handler) WebsocketRequestStatsReplyMessage {
	timeDiff := uint64(time.Since(handler.LastSnapshot.Timestamp).Seconds())
	handler.LastSnapshot.Timestamp = time.Now()
	message := WebsocketRequestStatsReplyMessage{
		CPU:   CPUUsage{},
		Disks: []DiskUsage{},
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, memory := GetMemoryDetails()
		message.Memory = memory
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		rawCpuUsage := GetCPUUsage()
		rawCpuTotal := float32(rawCpuUsage.Total - handler.LastSnapshot.RawCPUUsage.Total)
		message.CPU = CPUUsage{
			User:   (float32(rawCpuUsage.User-handler.LastSnapshot.RawCPUUsage.User) / rawCpuTotal) * 100,
			System: (float32(rawCpuUsage.System-handler.LastSnapshot.RawCPUUsage.System) / rawCpuTotal) * 100,
		}
		handler.LastSnapshot.RawCPUUsage = rawCpuUsage
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		networkUsage := GetNetworkUsage()
		message.Network = NetworkUsage{
			RX: MathMinUint(networkUsage.RX-handler.LastSnapshot.NetworkUsage.RX, 0),
			TX: MathMinUint(networkUsage.TX-handler.LastSnapshot.NetworkUsage.TX, 0),
		}
		handler.LastSnapshot.NetworkUsage = networkUsage
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		containerUsagesSnapshot := GetContainerUsages(handler)
		var containerUsages []ContainerUsage = []ContainerUsage{}
		for _, containerUsageSnapshot := range containerUsagesSnapshot {
			lastContainerUsageIndex := -1
			for i, tempContainerUsage := range handler.LastSnapshot.ContainerUsages {
				if containerUsageSnapshot.Parent == tempContainerUsage.Parent {
					lastContainerUsageIndex = i
				}
			}
			if lastContainerUsageIndex == -1 {
				continue
			}
			containerUsage := ContainerUsage{
				Parent: containerUsageSnapshot.Parent,
				RX:     MathMinUint(containerUsageSnapshot.RX-handler.LastSnapshot.ContainerUsages[lastContainerUsageIndex].RX, 0),
				TX:     MathMinUint(containerUsageSnapshot.TX-handler.LastSnapshot.ContainerUsages[lastContainerUsageIndex].TX, 0),
				CPU:    containerUsageSnapshot.CPU,
				Memory: containerUsageSnapshot.Memory,
				Read:   MathMinUint(containerUsageSnapshot.Read-handler.LastSnapshot.ContainerUsages[lastContainerUsageIndex].Read, 0) / uint64(timeDiff),
				Write:  MathMinUint(containerUsageSnapshot.Write-handler.LastSnapshot.ContainerUsages[lastContainerUsageIndex].Write, 0) / uint64(timeDiff),
			}
			containerUsages = append(containerUsages, containerUsage)
		}
		message.Containers = containerUsages
		handler.LastSnapshot.ContainerUsages = containerUsagesSnapshot
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		disks := GetDisks()
		diskUsagesSnapshot := GetDiskUsages()
		var diskUsages []DiskUsage = []DiskUsage{}
		for _, rawDiskUsage := range diskUsagesSnapshot {
			lastDiskUsageIndex := -1
			for i, diskUsage := range handler.LastSnapshot.RawDiskUsages {
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
			readsDiff := rawDiskUsage.Reads - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].Reads
			if readsDiff <= 0 {
				readsDiff = 1
			}
			writesDiff := rawDiskUsage.Writes - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].Writes
			if writesDiff <= 0 {
				writesDiff = 1
			}

			diskUsages = append(diskUsages, DiskUsage{
				Parent:       disks[matchingDiskIndex].ID,
				Read:         ((rawDiskUsage.ReadSectors - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].ReadSectors) * 512) / timeDiff,
				Write:        ((rawDiskUsage.WriteSectors - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].WriteSectors) * 512) / timeDiff,
				ReadLatency:  ((rawDiskUsage.ReadTime - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].ReadTime) / readsDiff),
				WriteLatency: ((rawDiskUsage.WriteTime - handler.LastSnapshot.RawDiskUsages[lastDiskUsageIndex].WriteTime) / writesDiff),
			})
			// SleepyLogLn("Adding... (name: %s, id: %s, read: %v, write: %v, timeDiff: %v)", disks[matchingDiskIndex].Name, disks[matchingDiskIndex].ID, diskUsage.Read, diskUsage.Write, timeDiff)
		}
		message.Disks = diskUsages
		handler.LastSnapshot.RawDiskUsages = diskUsagesSnapshot
	}()
	wg.Wait()

	return message
}
