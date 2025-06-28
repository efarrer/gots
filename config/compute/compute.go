package compute

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// ComputeGoCompilePath returns the path to compile the Go executable
func ComputeGoCompilePath(execName *string) func() (string, error) {
	return func() (string, error) {
		if execName == nil {
			return "", errors.New("ExecName not set")
		}
		compilePath := "./cmd/" + *execName
		fileInfo, err := os.Stat(compilePath)
		if err != nil {
			return "", err
		}
		if !fileInfo.IsDir() {
			return "", err
		}

		return compilePath, nil
	}
}

// GetCmd checks for a directory structure of ./cmd/<name> and if so it returns <name>. If not it returns ""
func GetCmd() (string, error) {
	dirPath := "./cmd/"
	fileInfo, err := os.Stat(dirPath)
	if err != nil {
		return "", err
	}

	if !fileInfo.IsDir() {
		return "", errors.New(fmt.Sprintf("%s is not a directory", fileInfo.Name()))
	}

	var cmd string
	err = filepath.WalkDir(dirPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && path != dirPath { // Exclude the directory itself
			cmd = filepath.Base(path)
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	return cmd, nil
}

var Getwd = os.Getwd
