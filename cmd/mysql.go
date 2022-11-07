package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func CreateBackup(handler *Handler, database string, args ...string) (string, error) {
	executable := GetMySQLDump()
	if executable == "" {
		return "", errors.New("could not find 'mysqldump'")
	}
	dumpPath := filepath.Join(handler.Directory, "temp")
	os.MkdirAll(dumpPath, 0755)

	for _, credentials := range handler.Credentials.Databases {
		for _, localDatabase := range credentials.Databases {
			if localDatabase.ID == database {
				path := filepath.Join(dumpPath, localDatabase.Name+".sql")
				cmdArgs := []string{"-h", credentials.Host, "-P", credentials.Port, "-u", credentials.Username, "-p" + credentials.Password}
				cmdArgs = append(cmdArgs, args...)
				cmdArgs = append(cmdArgs, localDatabase.Name, "--result-file="+path)
				_, err := exec.Command(executable, cmdArgs...).Output()
				if err != nil {
					SleepyErrorLn("%v", cmdArgs)
					return "", err
				}

				return path, nil
			}
		}
	}

	return "", errors.New("database isn't specified in the config")
}

func GetMySQLDump() string {
	return GetMySQLDumpSystem(runtime.GOOS)
}

func GetMySQLDumpSystem(system string) string {
	err := exec.Command("mysqldump").Run()
	if err != nil && strings.Contains(err.Error(), "not found") {
		err := exec.Command(fmt.Sprintf("tools/%s/mysqldump", system)).Run()
		if err != nil && strings.Contains(err.Error(), "not found") {
			return ""
		}

		return fmt.Sprintf("tools/%s/mysqldump", system)
	}

	return "mysqldump"
}
