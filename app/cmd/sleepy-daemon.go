package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Directory     string
	Config        Config
	StatsSnapshot HandlerStatsSnapshot
	WS            *websocket.Conn
	Session       *Session
	LogManager    DaemonLogManager
}

type HandlerStatsSnapshot struct {
	Timestamp          time.Time
	RawCPUUsage        CPUUsageRaw
	LinuxRawDiskUsages []DiskUsageLinuxRaw
	NetworkUsage       NetworkUsage
	ContainerUsages    []ContainerUsage
}

func CreateHandler(configName string) Handler {
	var handler Handler
	handler.Directory, _ = os.Getwd()
	handler.Config = NewConfig()
	configRaw, err := os.ReadFile(filepath.Join(handler.Directory, "config", configName))
	if err != nil {
		SleepyErrorLn("Failed to read config! Make sure you launched the daemon from the correct folder! (%s)", err.Error())
		closeDaemon(&handler)
		return handler
	}
	err = json.Unmarshal(configRaw, &handler.Config)
	if err != nil {
		SleepyErrorLn("Failed to parse config! (%s)", err.Error())
		closeDaemon(&handler)
		return handler
	}
	handler.StatsSnapshot.Timestamp = time.Now()
	handler.StatsSnapshot.RawCPUUsage = GetCPUUsage()
	if runtime.GOOS == "linux" {
		handler.StatsSnapshot.LinuxRawDiskUsages = GetDiskUsagesLinux()
	}
	handler.StatsSnapshot.NetworkUsage = GetNetworkUsage()
	handler.StatsSnapshot.ContainerUsages = GetContainerUsages()
	handler.LogManager.Containers = make(map[string]DaemonLogItem)

	return handler
}

func main() {
	// Flags Setup
	flagConfigName := flag.String("config", "default.json", "a config file")
	flagVersion := flag.Bool("v", false, "prints current daemon version")
	flag.Parse()
	if *flagVersion {
		fmt.Printf("sleepy-daemon v%s\n", DaemonVersion)
		os.Exit(0)
	}

	// Interrupt Setup
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Handler
	handler := CreateHandler(*flagConfigName)

	// Websocket
	var ws *websocket.Conn
	defer ws.Close()

	var wsLoop func()
	wsLoop = func() {
		ws := ConnectWebsocket(&handler)
		if ws == nil {
			time.Sleep(time.Second * time.Duration(handler.Config.ReconnectTimeout))
			go wsLoop()
			return
		}
		handler.WS = ws
		ProcessWebsocket(&handler, ws)
		handler.WS = nil
		handler.Session = nil
		time.Sleep(time.Second * time.Duration(handler.Config.ReconnectTimeout))
		go wsLoop()
	}
	go wsLoop()

	// Exit
	for {
		select {
		case <-interrupt:
			closeDaemon(&handler)
		}
	}
}

func closeDaemon(handler *Handler) {
	closeDaemonNoExit(handler)
	os.Exit(0)
}

func closeDaemonNoExit(handler *Handler) {
	if handler.WS != nil {
		// Cleanly close the connection by sending a close message and then
		// waiting (with timeout) for the server to close the connection.
		err := handler.WS.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			SleepyWarnLn("write close: %s", err.Error())
			return
		}
		SleepyLogLn("Closed connection!")
	}
}
