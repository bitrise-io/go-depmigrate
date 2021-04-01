package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/bitirse-io/go-mod-update/depmigrate"
	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-io/stepman/models"
	"gopkg.in/yaml.v2"
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

	stepYMLPath := filepath.Join(projectDir, "step.yml")
	stepYML, err := os.ReadFile(stepYMLPath)
	if err != nil {
		failf("failed to read file (%s): %v", stepYMLPath, err)
	}

	var stepModel models.StepModel
	if err := yaml.Unmarshal(stepYML, &stepModel); err != nil {
		failf("failed to unmarshal step.yml: %v", err)
	}

	if stepModel.Toolkit != nil && stepModel.Toolkit.Go != nil {
		if migrator.IsGoPathModeStep() {
			if err := migrator.Migrate("go", "", stepModel.Toolkit.Go.PackageName); err != nil {
				failf("Failed to update to go modules: ", err)
			}

			for _, file := range []string{"Gopkg.lock", "Gopkg.toml"} {
				if err := os.Remove(filepath.Join(projectDir, file)); err != nil {
					log.Warnf("failed to remove file (%s): %v", file, err)
				}
			}

			if err := os.RemoveAll(filepath.Join(projectDir, "Godeps")); err != nil {
				log.Warnf("failed to remove Godeps: %v", err)
			}

			if err := buildInTmpDir(projectDir); err != nil {
				failf("%v", err)
			}
		}
	}
}

func buildInTmpDir(projectDir string) error {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return fmt.Errorf("failed to create temp dir: %v", err)
	}

	if err := command.CopyDir(projectDir, tmpDir, true); err != nil {
		return fmt.Errorf("failed to copy directory: %v", err)
	}

	buildCmd := command.New("go", "build").SetDir(tmpDir).SetStderr(os.Stderr).SetStdout(os.Stdout)
	log.Infof("$ %s", buildCmd.PrintableCommandArgs())
	if err := buildCmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			failf("go build command failed: %v", err)
		}

		failf("failed to run go build: %v", err)
	}

	if err := os.RemoveAll(tmpDir); err != nil {
		return fmt.Errorf("failed to remove tmp dir: %v", err)
	}

	return nil
}
