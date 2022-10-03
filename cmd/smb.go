package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

func RebuildSmbConfig(handler *Handler, config string) {
	smbPath := filepath.Join(handler.Directory, "containers", "sleepy-smb")
	smbDockerPath := filepath.Join(smbPath, "docker-compose.yml")

	if _, err := os.Stat(smbDockerPath); err == nil {
		dockerCmd := exec.Command("docker-compose", "rm", "-f", "-s", "-v")
		dockerCmd.Dir = smbPath
		dockerStdout, err := dockerCmd.Output()
		if err != nil {
			SleepyWarnLn("Failed to stop previous SMB containers! (%s)", err.Error())
			SleepyWarnLn(string(dockerStdout))
			return
		}
	}

	for _, user := range handler.Credentials.Smb {
		config = strings.ReplaceAll(config, fmt.Sprintf("%%SMB_USER_%s_PASSWORD%%", user.ID), user.Password)
	}
	re, _ := regexp.Compile("%SMB_USER_(.*?)_PASSWORD%")
	missingUsers := re.FindAllStringSubmatch(config, -1)
	for i := range missingUsers {
		SleepyWarnLn("Missing credentials for SMB user! (id: %s)", missingUsers[i][1])
	}
	if len(missingUsers) > 0 {
		return
	}

	os.MkdirAll(smbPath, 0755)
	err := os.WriteFile(smbDockerPath, []byte(config), 0644)
	if err != nil {
		SleepyWarnLn("Failed to write SMB docker-compose.yml! (%s)", err.Error())
		return
	}

	dockerCmd := exec.Command("docker-compose", "up", "-d")
	dockerCmd.Dir = smbPath
	dockerStdout, err := dockerCmd.Output()
	if err != nil {
		SleepyWarnLn("Failed to start new SMB containers! (%s)", err.Error())
		SleepyWarnLn(string(dockerStdout))
		return
	}
}
