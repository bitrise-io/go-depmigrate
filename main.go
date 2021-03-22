package main

import (
	"os"
	"strings"

	"github.com/bitirse-io/go-mod-update/depmigrate"
	"github.com/bitrise-io/go-utils/log"
)

func failf(format string, v ...interface{}) {
	log.Errorf(format, v...)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		failf(`Usage:
modupdate projectPath`)
	}

	projectDir := strings.TrimSpace(os.Args[1])
	if projectDir == "" {
		failf("Empty project path specified.")
	}

	migrator, err := depmigrate.NewGoModMigrator(projectDir)
	if err != nil {
		failf("%s", err)
	}

	if migrator.ShouldMigrate() {
		if err := migrator.UpdateProjectToGoModules(); err != nil {
			failf("Failed to update to go modules: ", err)
		}
	}
}
