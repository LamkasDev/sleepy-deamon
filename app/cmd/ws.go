package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jwalton/gchalk"
)

const (
	WebsocketMessageTypeAuth        string = "DAEMON_AUTH"
	WebsocketMessageTypeAuthSuccess string = "DAEMON_AUTH_SUCCESS"
	WebsocketMessageTypeAuthFailure string = "DAEMON_AUTH_FAILURE"

	WebsocketMessageTypeRequestRefresh        string = "DAEMON_REQUEST_REFRESH"
	WebsocketMessageTypeRequestRefreshReply   string = "DAEMON_REQUEST_REFRESH_REPLY"
	WebsocketMessageTypeRequestDatabaseBackup string = "DAEMON_REQUEST_DATABASE_BACKUP"

	WebsocketMessageTypeRequestStats      string = "DAEMON_REQUEST_STATS"
	WebsocketMessageTypeRequestStatsReply string = "DAEMON_REQUEST_STATS_REPLY"
)

const (
	WebsocketAuthFailureWrongToken      string = "WRONG_TOKEN"
	WebsocketAuthFailureVersionMismatch string = "VERSION_MISMATCH"
)

type WebsocketMessage struct {
	Type string `json:"type"`
}

type WebsocketAuthMessage struct {
	Type    string `json:"type"`
	Token   string `json:"token"`
	Version string `json:"version"`
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

type WebsocketRequestRefreshReplyMessage struct {
	Type       string      `json:"type"`
	Disks      []Disk      `json:"disks"`
	Containers []Container `json:"containers"`
}

type WebsocketRequestDatabaseBackupMessage struct {
	Type     string `json:"type"`
	Database string `json:"database"`
}

type WebsocketRequestStatsReplyMessage struct {
	Type    string       `json:"type"`
	CPU     CPUUsage     `json:"cpu"`
	Memory  MemoryUsage  `json:"memory"`
	Disks   []DiskUsage  `json:"disks"`
	Network NetworkUsage `json:"network"`
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
		Type:    WebsocketMessageTypeAuth,
		Token:   handler.Config.Token,
		Version: DaemonVersion,
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
			SleepyLogLn("Logged in as %s! (id: %s)", message.Name, message.ID)
			break
		case WebsocketMessageTypeAuthFailure:
			var message WebsocketAuthFailureMessage
			_ = json.Unmarshal(messageRaw, &message)
			switch message.Reason {
			case WebsocketAuthFailureWrongToken:
				SleepyErrorLn("Incorrect token! Closing the deamon...")
				closeDaemon(handler)
				break
			case WebsocketAuthFailureVersionMismatch:
				var message WebsocketAuthFailureVersionMismatchMessage
				_ = json.Unmarshal(messageRaw, &message)
				SleepyWarnLn("Version mismatch! Current version %s is not needed %s! Updating...", gchalk.Red(DaemonVersion), gchalk.Green(message.Version))
				err := Update(handler, message.Version)
				if err != nil {
					return err
				}
				break
			default:
				return errors.New(fmt.Sprintf("failed to auth: %s", message.Reason))
			}
			break
		case WebsocketMessageTypeRequestRefresh:
			requestRefreshReplyMessage := WebsocketRequestRefreshReplyMessage{
				Type:       WebsocketMessageTypeRequestRefreshReply,
				Disks:      GetDisks(),
				Containers: GetContainers(),
			}
			ws.WriteJSON(requestRefreshReplyMessage)
			break
		case WebsocketMessageTypeRequestDatabaseBackup:
			var message WebsocketRequestDatabaseBackupMessage
			_ = json.Unmarshal(messageRaw, &message)
			err := CreateBackup(handler, message.Database)
			if err != nil {
				SleepyWarnLn("Failed to create a database backup! (%s)", err.Error())
				continue
			}

			uploadFileData := UploadFileBackupDatabaseData{
				Type:     UploadFileDataBackupDatabase,
				Database: message.Database,
			}
			err = UploadFile(handler, filepath.Join(handler.Directory, "dump", fmt.Sprintf("%s.sql", message.Database)), uploadFileData)
			if err != nil {
				SleepyWarnLn("Failed to upload database backup! (%s)", err.Error())
			}
			break
		case WebsocketMessageTypeRequestStats:
			requestStatsReplyMessage := GetStatsMessage(handler)
			requestStatsReplyMessage.Type = WebsocketMessageTypeRequestStatsReply
			ws.WriteJSON(requestStatsReplyMessage)
			break
		}
	}
}

func GetStatsMessage(handler *Handler) WebsocketRequestStatsReplyMessage {
	timeDiff := uint64(time.Now().Sub(handler.StatsSnapshot.Timestamp).Seconds())
	handler.StatsSnapshot.Timestamp = time.Now()
	message := WebsocketRequestStatsReplyMessage{
		CPU:    CPUUsage{},
		Memory: GetMemoryUsage(),
		Disks:  []DiskUsage{},
	}

	rawCpuUsage := GetCPUUsage()
	rawCpuTotal := float32(rawCpuUsage.Total - handler.StatsSnapshot.RawCPUUsage.Total)
	message.CPU = CPUUsage{
		Total: (float32(rawCpuUsage.User-handler.StatsSnapshot.RawCPUUsage.User) + float32(rawCpuUsage.System-handler.StatsSnapshot.RawCPUUsage.System)) / rawCpuTotal * 100,
	}
	handler.StatsSnapshot.RawCPUUsage = rawCpuUsage

	networkUsage := GetNetworkUsage()
	message.Network = NetworkUsage{
		RX: (networkUsage.RX - handler.StatsSnapshot.NetworkUsage.RX) / timeDiff,
		TX: (networkUsage.TX - handler.StatsSnapshot.NetworkUsage.TX) / timeDiff,
	}
	if message.Network.RX < 0 {
		message.Network.RX = 0
	}
	if message.Network.TX < 0 {
		message.Network.TX = 0
	}
	handler.StatsSnapshot.NetworkUsage = networkUsage

	switch runtime.GOOS {
	case "linux":
		var diskUsages []DiskUsage
		disks := GetDisks()
		lastRawDiskUsages := handler.StatsSnapshot.LinuxRawDiskUsages
		rawDiskUsages := GetDiskUsagesLinux()
		for _, rawDiskUsage := range rawDiskUsages {
			lastRawDiskUsageIndex := -1
			for i, diskUsage := range lastRawDiskUsages {
				if diskUsage.Name == rawDiskUsage.Name {
					lastRawDiskUsageIndex = i
				}
			}
			if lastRawDiskUsageIndex == -1 {
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
			readsDiff := rawDiskUsage.Reads - lastRawDiskUsages[lastRawDiskUsageIndex].Reads
			if readsDiff <= 0 {
				readsDiff = 1
			}
			writesDiff := rawDiskUsage.Writes - lastRawDiskUsages[lastRawDiskUsageIndex].Writes
			if writesDiff <= 0 {
				writesDiff = 1
			}

			diskUsage := DiskUsage{
				Parent:       disks[matchingDiskIndex].ID,
				Read:         ((rawDiskUsage.ReadSectors - lastRawDiskUsages[lastRawDiskUsageIndex].ReadSectors) * 512) / timeDiff,
				Write:        ((rawDiskUsage.WriteSectors - lastRawDiskUsages[lastRawDiskUsageIndex].WriteSectors) * 512) / timeDiff,
				ReadLatency:  ((rawDiskUsage.ReadTime - lastRawDiskUsages[lastRawDiskUsageIndex].ReadTime) / readsDiff),
				WriteLatency: ((rawDiskUsage.WriteTime - lastRawDiskUsages[lastRawDiskUsageIndex].WriteTime) / writesDiff),
			}
			diskUsages = append(diskUsages, diskUsage)
			// SleepyLogLn("Adding... (name: %s, id: %s, read: %v, write: %v, timeDiff: %v)", disks[matchingDiskIndex].Name, disks[matchingDiskIndex].ID, diskUsage.Read, diskUsage.Write, timeDiff)
		}
		message.Disks = diskUsages
		handler.StatsSnapshot.LinuxRawDiskUsages = rawDiskUsages
		break
	}

	return message
}
