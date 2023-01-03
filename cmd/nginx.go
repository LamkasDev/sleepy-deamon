package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func RebuildNginxConfig(handler *Handler, message WebsocketBuildNginxConfigMessage) {
	nginxPath := filepath.Join(handler.Directory, "containers", "sleepy-nginx")
	nginxDockerComposePath := filepath.Join(nginxPath, "docker-compose.yml")
	nginxDockerfilePath := filepath.Join(nginxPath, "Dockerfile")
	nginxConfigPath := filepath.Join(nginxPath, "nginx.conf")
	nginxLivePath := filepath.Join(nginxPath, "live")
	nginxSiteConfigsPath := filepath.Join(nginxPath, "conf.d")

	if _, err := os.Stat(nginxDockerComposePath); err == nil {
		dockerCmd := exec.Command("docker-compose", "rm", "-f", "-s", "-v")
		dockerCmd.Dir = nginxPath
		dockerStdout, err := dockerCmd.Output()
		if err != nil {
			SleepyWarnLn("Failed to stop previous NGINX containers! (%s)", err.Error())
			SleepyWarnLn(string(dockerStdout))
		}
	}

	os.MkdirAll(nginxPath, 0755)
	os.MkdirAll(nginxLivePath, 0755)
	err := os.WriteFile(nginxDockerComposePath, []byte(message.Config), 0644)
	if err != nil {
		SleepyWarnLn("Failed to write NGINX docker-compose.yml! (%s)", err.Error())
		return
	}
	err = os.WriteFile(nginxDockerfilePath, []byte(message.Dockerfile), 0644)
	if err != nil {
		SleepyWarnLn("Failed to write NGINX Dockerfile! (%s)", err.Error())
		return
	}
	err = os.WriteFile(nginxConfigPath, []byte(message.NginxConfig), 0644)
	if err != nil {
		SleepyWarnLn("Failed to write NGINX nginx.conf! (%s)", err.Error())
		return
	}

	os.RemoveAll(nginxSiteConfigsPath)
	os.MkdirAll(nginxSiteConfigsPath, 0755)
	for _, config := range message.Servers {
		nginxSiteConfigPath := filepath.Join(nginxSiteConfigsPath, fmt.Sprintf("%s.conf", config.Name))
		err = os.WriteFile(nginxSiteConfigPath, []byte(config.Config), 0644)
		if err != nil {
			SleepyWarnLn("Failed to write NGINX %s! (%s)", config.Name, err.Error())
			return
		}
	}

	for _, config := range message.Servers {
		nginxCertificatePath := filepath.Join(nginxLivePath, config.Ssl, "fullchain.pem")
		nginxKeyPath := filepath.Join(nginxLivePath, config.Ssl, "privkey.pem")
		if _, err = os.Stat(nginxCertificatePath); err != nil {
			SleepyWarnLn("Missing certificate for server! (name: %s)", config.Ssl)
			return
		}
		if _, err = os.Stat(nginxKeyPath); err != nil {
			SleepyWarnLn("Missing key for server! (name: %s)", config.Ssl)
			return
		}
	}

	dockerCmd := exec.Command("docker-compose", "build")
	dockerCmd.Dir = nginxPath
	dockerStdout, err := dockerCmd.Output()
	if err != nil {
		SleepyWarnLn("Failed to build new NGINX containers! (%s)", err.Error())
		SleepyWarnLn(string(dockerStdout))
		return
	}

	for _, network := range message.Networks {
		_, _ = exec.Command("docker", "network", "create", network).Output()
	}

	dockerCmd = exec.Command("docker-compose", "up", "-d")
	dockerCmd.Dir = nginxPath
	dockerStdout, err = dockerCmd.Output()
	if err != nil {
		SleepyWarnLn("Failed to start new NGINX containers! (%s)", err.Error())
		SleepyWarnLn(string(dockerStdout))
		return
	}
}
