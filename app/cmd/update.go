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
	SleepyLog("Downloading %s (%v%%)...", gchalk.Yellow(wc.Version), (wc.Downloaded/wc.Total)*100)
	return n, nil
}

func Update(handler *Handler, version string) error {
	os.MkdirAll(filepath.Join(handler.Directory, "temp"), 0755)

	// Main Daemon
	var url string
	if handler.Config.Https {
		url = fmt.Sprintf("https://%s/daemons/%s.zip", handler.Config.DataHost, version)
	} else {
		url = fmt.Sprintf("http://%s/daemons/%s.zip", handler.Config.DataHost, version)
	}
	err := Download(handler, url, filepath.Join(handler.Directory, "temp", "daemon.zip"), fmt.Sprintf("daemon-%s", version))
	if err != nil {
		return err
	}

	// Root Daemon Files
	if handler.Config.Https {
		url = fmt.Sprintf("https://%s/daemons/%s-root.zip", handler.Config.DataHost, version)
	} else {
		url = fmt.Sprintf("http://%s/daemons/%s-root.zip", handler.Config.DataHost, version)
	}
	err = Download(handler, url, filepath.Join(handler.Directory, "temp", "daemon-root.zip"), fmt.Sprintf("daemon-root-%s", version))
	if err != nil {
		return err
	}

	SleepyLogLn("Deleting old files...")
	dir, err := os.ReadDir(handler.Directory)
	if err != nil {
		return err
	}
	for _, f := range dir {
		if !f.IsDir() || f.Name() == "misc" || f.Name() == "tools" {
			os.RemoveAll(filepath.Join(handler.Directory, f.Name()))
		}
	}

	SleepyLogLn("Extracting archives...")
	Unzip(filepath.Join(handler.Directory, "temp", "daemon.zip"), filepath.Join(handler.Directory, version))
	Unzip(filepath.Join(handler.Directory, "temp", "daemon-root.zip"), handler.Directory)

	SleepyLogLn("Setting current version to %s...", gchalk.Yellow(version))
	err = os.WriteFile("current_version.txt", []byte(version), 0644)
	if err != nil {
		return err
	}

	SleepyLogLn("Closing current daemon...")
	closeDaemonNoExit(handler)

	SleepyLogLn("Launching new daemon with redirected output...")
	SleepyLogLn(gchalk.Magenta("-> NEW DEAMON OUTPUT"))
	daemonCmd := exec.Command(filepath.Join(version, "bin", "sleepy-daemon"), "-config=windows.json")
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

	SleepyLog("Downloading...")
	fileBody := io.TeeReader(resp.Body, &WriteCounter{Version: version, Total: int64(size)})
	_, err = io.Copy(out, fileBody)
	if err != nil {
		return err
	}
	fmt.Printf("\n")
	return nil
}
