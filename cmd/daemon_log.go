package main

import "os/exec"

type DaemonLogManager struct {
	Containers map[string]DaemonLogItem
}
type DaemonLogItem struct {
	Command *exec.Cmd
}
