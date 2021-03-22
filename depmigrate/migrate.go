package depmigrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/bitrise-io/go-utils/command"
	"github.com/bitrise-io/go-utils/errorutil"
	"github.com/bitrise-io/go-utils/log"
)

type GoModMigrator struct {
	projectDir string
}

func NewGoModMigrator(projectDir string) (*GoModMigrator, error) {
	absPath, err := filepath.Abs(projectDir)
	if err != nil {
		return nil, err
	}

	fileInfo, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project path: %v", err)
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("not a directory (%s)", absPath)
	}

	return &GoModMigrator{projectDir: projectDir}, nil
}

func (m GoModMigrator) ShouldMigrate() bool {
	goModPath := filepath.Join(m.projectDir, "go.mod")
	_, err := os.Stat(goModPath)

	return err == nil
}

func (m GoModMigrator) UpdateProjectToGoModules() error {
	cmds := []*command.Model{
		command.New("go", "mod", "init"),
		command.New("go", "mod", "tidy"),
		command.New("go", "mod", "vendor"),
	}

	for _, cmd := range cmds {
		cmd.SetDir(m.projectDir).SetStdout(os.Stdout).SetStderr(os.Stderr)

		fmt.Println()
		log.Infof("$ %s", cmd.PrintableCommandArgs())

		if err := cmd.Run(); err != nil {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("command exited with nonzero status: %v", err)
			}

			return fmt.Errorf("failed to run command: %v", err)
		}
	}

	return nil
}
