package main

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func CreateBackup(handler *Handler, database string) error {
	executable := GetMySQLDump()
	if executable == "" {
		return errors.New("Could not find 'mysqldump'!")
	}

	for _, credentials := range handler.Config.DatabaseCredentials {
		for _, db := range credentials.Databases {
			if db == database {
				err := exec.Command(executable, "-h", credentials.Host, "-P", credentials.Port, "-u", credentials.Username, "-p"+credentials.Password, db, "--result-file="+filepath.Join(handler.Directory, "dump", db+".sql")).Run()
				if err != nil {
					return err
				}

				return nil
			}
		}
	}

	return errors.New("Database isn't specified in the config!")
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
