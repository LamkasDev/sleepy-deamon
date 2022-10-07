package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	"github.com/jwalton/gchalk"
)

type WriteCounter struct {
	Version    string
	Downloaded int64
	Total      int64
}

func (wc *WriteCounter) Write(p []byte) (int, error) {
	n := len(p)
	wc.Downloaded += int64(n)
	fmt.Print("\r")
	SleepyLog("Pulling %s (%v%%)...", gchalk.Yellow(wc.Version), (wc.Downloaded/wc.Total)*100)
	return n, nil
}

func Update(handler *Handler, version string) error {
	os.MkdirAll(filepath.Join(handler.Directory, "temp"), 0755)

	// Download
	SleepyLogLn("Downloading archive...")
	url := fmt.Sprintf("https://%s/daemons/%s.zip", handler.Config.DataHost, version)
	err := Download(handler, url, filepath.Join(handler.Directory, "temp", "daemon.zip"), version+".zip")
	if err != nil {
		SleepyWarnLn("Failed to download archive! (%s)", err.Error())
		return err
	}

	// Delete existing directory
	SleepyLogLn("Deleting existing directory (if exists)...")
	os.RemoveAll(filepath.Join(handler.Directory, version))

	// Extract
	SleepyLogLn("Extracting archive...")
	Unzip(filepath.Join(handler.Directory, "temp", "daemon.zip"), filepath.Join(handler.Directory, version))

	// Change file permissions
	SleepyLogLn("Changing file permissions...")
	_, err = exec.Command("chmod", "-R", "a+rx", filepath.Join(handler.Directory, version, "scripts")).Output()
	if err != nil {
		SleepyWarnLn("Failed to change file permissions! (%s)", err.Error())
		return err
	}

	// Build
	SleepyLogLn("Building daemon...")
	buildCmd := exec.Command("/bin/bash", "build-linux.sh")
	buildCmd.Dir = filepath.Join(handler.Directory, version, "scripts")
	_, err = buildCmd.Output()
	if err != nil {
		SleepyWarnLn("Failed to get build the daemon! (%s)", err.Error())
		return err
	}

	// Close
	SleepyLogLn("Closing current daemon...")
	closeDaemonNoExit(handler)

	// Launch
	SleepyLogLn("Launching new daemon with redirected output...")
	SleepyLogLn(gchalk.Magenta("-> NEW DEAMON OUTPUT"))
	daemonCmd := exec.Command(filepath.Join(version, "bin", "sleepy-daemon"), "-config=current.json")
	daemonCmd.Dir = handler.Directory
	daemonCmd.Stdout = os.Stdout
	daemonCmd.Stdout = os.Stdout
	err = daemonCmd.Run()
	if err != nil {
		return err
	}

	os.Exit(0)
	return nil
}

func Download(handler *Handler, url string, path string, version string) error {
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	size, _ := strconv.Atoi(resp.Header.Get("Content-Length"))

	SleepyLog("Pulling...")
	fileBody := io.TeeReader(resp.Body, &WriteCounter{Version: version, Total: int64(size)})
	_, err = io.Copy(out, fileBody)
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}
