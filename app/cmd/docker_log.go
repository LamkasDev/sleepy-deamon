package main

import (
	"bufio"
	"os/exec"
)

func ConnectContainerLogger(handler *Handler, container WebsocketConnectContainerContainer) {
	cmd := exec.Command("docker", "logs", container.Name, "--follow", "--tail", "50")
	pipe, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		SleepyErrorLn("Failed to connect container logger! (%s)", err.Error())
		return
	}

	go func() {
		defer cmd.Wait()
		defer SleepyLogLn("Disconnected container logger! (id: %s)", container.ID)
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			line := scanner.Text()
			logMessage := WebsocketContainerLogMessageMessage{
				Type:      WebsocketMessageTypeContainerLogMessage,
				Container: container.ID,
				Message:   line,
			}

			handler.WS.WriteJSON(logMessage)
		}
	}()
	handler.LogManager.Containers[container.ID] = DaemonLogItem{
		Command: cmd,
	}
	SleepyLogLn("Connected container logger! (id: %s)", container.ID)
}

func DisconnectContainerLogger(handler *Handler, container string) {
	item, ok := handler.LogManager.Containers[container]
	if !ok {
		SleepyWarnLn("Failed to disconnect container logger! (%s)", "not found")
		return
	}

	item.Command.Process.Kill()
}
