package main

import (
	"bufio"
	"io"
	"os/exec"
	"strconv"
)

func RequestContainerLog(handler *Handler, container Container, task string) {
	path := ConvertDockerPath(handler, container.Log)
	taskProgressMessage := WebsocketTaskProgressMessage{
		Type:   WebsocketMessageTypeTaskProgress,
		ID:     task,
		Status: TaskStatusRunning,
	}
	uploadFileData := UploadFileContainerLogData{
		Type:      UploadFileDataContainerLog,
		Container: container.ID,
		Task:      task,
	}
	err := UploadFile(handler, path, uploadFileData)
	if err != nil {
		SleepyWarnLn("Failed to upload container log! (%s)", err.Error())
		taskProgressMessage.Status = TaskStatusFailed
		SendWebsocketMessage(handler, taskProgressMessage)
		return
	}
	taskProgressMessage.Status = TaskStatusFinished
	taskProgressMessage.Progress = 100
	SendWebsocketMessage(handler, taskProgressMessage)
}

func ConnectContainerLogger(handler *Handler, container WebsocketConnectContainerContainer, options WebsocketConnectContainerOptions) {
	var cmd *exec.Cmd = nil
	if options.Project {
		cmd = exec.Command("docker-compose", "logs", "--follow", "--tail", strconv.Itoa(int(options.Tail)))
		cmd.Dir = *container.Path
	} else {
		cmd = exec.Command("docker", "logs", container.Name, "--follow", "--tail", strconv.Itoa(int(options.Tail)))
	}

	pipe, _ := cmd.StdoutPipe()
	err := cmd.Start()
	if err != nil {
		SleepyErrorLn("Failed to connect container logger! (%s)", err.Error())
		return
	}
	ConnectContainerLoggerInternal(handler, container.ID, cmd, pipe)
}

func ConnectContainerLoggerInternal(handler *Handler, id string, cmd *exec.Cmd, pipe io.ReadCloser) {
	go func() {
		defer cmd.Wait()
		defer SleepyLogLn("Disconnected container logger! (id: %s)", id)
		scanner := bufio.NewScanner(pipe)
		for scanner.Scan() {
			line := scanner.Text()
			logMessage := WebsocketContainerLogMessageMessage{
				Type:    WebsocketMessageTypeContainerLogMessage,
				ID:      id,
				Message: line,
			}

			SendWebsocketMessage(handler, logMessage)
		}
	}()
	handler.LogManager.Containers[id] = DaemonLogItem{
		Command: cmd,
	}
	SleepyLogLn("Connected container logger! (id: %s)", id)
}

func DisconnectContainerLogger(handler *Handler, ID string) {
	item, ok := handler.LogManager.Containers[ID]
	if !ok {
		SleepyWarnLn("Failed to disconnect container logger! (%s)", "not found")
		return
	}

	item.Command.Process.Kill()
}
