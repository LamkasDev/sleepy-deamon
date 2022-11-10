package main

import (
	_ "embed"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Handler struct {
	Directory    string
	Config       Config
	Credentials  ConfigCredentials
	LastSnapshot HandlerSnapshot
	LastCache    HandlerCache
	WSMutex      *sync.Mutex
	WS           *websocket.Conn
	Session      *Session
	LogManager   DaemonLogManager
}

func ReadConfig(handler *Handler, name string, target any, def any) bool {
	path := filepath.Join(handler.Directory, "config", name)
	raw, err := os.ReadFile(path)
	if err != nil {
		SleepyWarnLn("%s not found, creating default one...", name)
		raw, _ = json.Marshal(def)
		os.WriteFile(path, raw, 0755)
	}
	err = json.Unmarshal(raw, target)
	if err != nil {
		SleepyErrorLn("Failed to parse %s! (%s)", name, err.Error())
		closeDaemon(handler)
		return false
	}

	return true
}

func CreateHandler(configName string) Handler {
	var handler Handler
	handler.Directory, _ = os.Getwd()
	handler.Config = NewConfig()
	handler.Credentials = NewConfigCredentials()
	os.MkdirAll(filepath.Join(handler.Directory, "config"), 0755)
	os.MkdirAll(filepath.Join(handler.Directory, "temp"), 0755)

	if !ReadConfig(&handler, configName, &handler.Config, NewConfig()) {
		return handler
	}
	if !ReadConfig(&handler, "credentials.json", &handler.Credentials, NewConfigCredentials()) {
		return handler
	}

	return handler
}

func main() {
	// Flags Setup
	flagConfigName := flag.String("config", "default.json", "a config file")
	flagVersion := flag.Bool("v", false, "prints current daemon version")
	flagDebug := flag.Bool("d", false, "runs in debug mode")
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

	// Setup profiling
	if *flagDebug {
		f, err := os.Create(filepath.Join(handler.Directory, "temp", "cpu.prof"))
		if err != nil {
			log.Fatal(err)
			return
		}
		pprof.StartCPUProfile(f)
		SleepyLogLn("Running in debug mode...")
	}

	// Websocket
	var ws *websocket.Conn
	defer ws.Close()

	// Websocket processsing
	var wsLoop func()
	wsLoop = func() {
		// Create mutex
		mutex := sync.Mutex{}
		handler.WSMutex = &mutex

		// Connect websocket to server
		ws := ConnectWebsocket(&handler)
		if ws == nil {
			time.Sleep(time.Second * time.Duration(handler.Config.ReconnectTimeout))
			go wsLoop()
			return
		}
		handler.WS = ws

		// Authenticate and process messages (blocking)
		AuthWebsocket(&handler)
		ProcessWebsocket(&handler, ws)

		// Something happened, so let's prepare for a fresh start
		handler.WSMutex = nil
		handler.WS = nil
		handler.Session = nil

		// After ReconnectTimeout passed, try again
		time.Sleep(time.Second * time.Duration(handler.Config.ReconnectTimeout))
		go wsLoop()
	}
	go wsLoop()

	// Wait for exit
	for {
		select {
		case <-interrupt:
			closeDaemon(&handler)
		}
	}
}

func closeDaemon(handler *Handler) {
	closeDaemonNoExit(handler)
	pprof.StopCPUProfile()
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
