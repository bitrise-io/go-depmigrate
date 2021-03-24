package depmigrate

import (
	"fmt"
	"io/ioutil"
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

func (m GoModMigrator) IsGoPathModeStep() bool {
	goModPath := filepath.Join(m.projectDir, "go.mod")
	_, err := os.Stat(goModPath)

	return err != nil
}

func (m GoModMigrator) UpdateProjectToGoModules(goBinaryPath, packageName string) error {
	initCmd := command.New(goBinaryPath, "mod", "init").SetDir(m.projectDir).SetStdout(os.Stdout).SetStderr(os.Stderr)
	if packageName != "" {
		initCmd = command.New(goBinaryPath, "mod", "init", packageName)
	}

	fmt.Println()
	log.Infof("$ %s in %s", initCmd.PrintableCommandArgs(), m.projectDir)

	if err := initCmd.Run(); err != nil {
		if errorutil.IsExitStatusError(err) {
			return fmt.Errorf("command exited with nonzero status: %v", err)
		}

		return fmt.Errorf("failed to run command: %v", err)
	}

	goModPath := filepath.Join(m.projectDir, "go.mod")
	// goModContents, err := ioutil.ReadFile(goModPath)
	// if err != nil {
	// 	return fmt.Errorf("failed to read file (%s): %v", goModPath, err)
	// }

	// Prevent error:
	// go mod tidy
	// go: github.com/Sirupsen/logrus@v1.4.2: parsing go.mod:
	// module declares its path as: github.com/sirupsen/logrus
	// 		but was required as: github.com/Sirupsen/logrus
	// newGoModContents := strings.ReplaceAll(string(goModContents), "github.com/Sirupsen/logrus", "github.com/sirupsen/logrus")
	// if err := ioutil.WriteFile(goModPath, []byte(newGoModContents), 0600); err != nil {
	// 	return fmt.Errorf("failed to write file (%s): %v", goModPath, err)
	// }

	cmds := []*command.Model{
		command.New(goBinaryPath, "mod", "tidy"),
		command.New(goBinaryPath, "mod", "vendor"),
	}

	for _, cmd := range cmds {
		cmd.SetDir(m.projectDir).SetStdout(os.Stdout).SetStderr(os.Stderr)

		lsCmd := command.New("tree", ".").SetDir(m.projectDir).SetStdout(os.Stdout).SetStderr(os.Stderr)
		if err := lsCmd.Run(); err != nil {
			return err
		}

		_, err := os.Stat(goModPath)
		if err != nil {
			log.Infof("go.mod does not exists: %s", err)
		} else {
			log.Infof("go.mod exist at %s", goModPath)
			data, err := ioutil.ReadFile(goModPath)
			if err != nil {
				return err
			}
			log.Infof("go.mod contents: %s", data)
		}

		fmt.Println()
		log.Infof("$ %s in %s", cmd.PrintableCommandArgs(), m.projectDir)
		if err := cmd.Run(); err != nil {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("command exited with nonzero status: %v", err)
			}

			return fmt.Errorf("failed to run command: %v", err)
		}
	}

	return nil
}

func (m GoModMigrator) Update() error {
	cmds := []*command.Model{
		command.New("go", "get", "-t", "-u", "./..."),
		command.New("go", "mod", "vendor"),
	}

	for _, cmd := range cmds {
		cmd.SetDir(m.projectDir).SetStdout(os.Stdout).SetStderr(os.Stderr)

		log.Infof("$ %s", cmd.PrintableCommandArgs())
		if err := cmd.Run(); err != nil {
			if errorutil.IsExitStatusError(err) {
				return fmt.Errorf("command failed: %v", err)
			}

			return fmt.Errorf("failed to run command: %v", err)
		}
	}

	return nil
}
