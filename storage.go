package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	applicationDirectoryName = "haitranslator"
	inputDirectoryName       = "input"
	outputDirectoryName      = "output"
	tokenUsageFileName       = "use_token.json"
)

type applicationDirectories struct {
	root       string
	input      string
	output     string
	tokenUsage string
}

func applicationDirectoriesFor(homeDir string) (applicationDirectories, error) {
	if homeDir == "" {
		return applicationDirectories{}, fmt.Errorf("thư mục home người dùng không hợp lệ")
	}

	root := filepath.Join(homeDir, applicationDirectoryName)
	return applicationDirectories{
		root:       root,
		input:      filepath.Join(root, inputDirectoryName),
		output:     filepath.Join(root, outputDirectoryName),
		tokenUsage: filepath.Join(root, tokenUsageFileName),
	}, nil
}

func defaultApplicationDirectories() (applicationDirectories, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return applicationDirectories{}, fmt.Errorf("không thể tìm thư mục home người dùng: %w", err)
	}
	return applicationDirectoriesFor(homeDir)
}

func ensureApplicationDirectories(homeDir string) (applicationDirectories, error) {
	directories, err := applicationDirectoriesFor(homeDir)
	if err != nil {
		return applicationDirectories{}, err
	}

	for _, directory := range []string{directories.root, directories.input, directories.output} {
		if err := os.MkdirAll(directory, 0o755); err != nil {
			return applicationDirectories{}, fmt.Errorf("không thể tạo thư mục %s: %w", directory, err)
		}
	}
	if _, err := newTokenUsageTracker(directories.tokenUsage); err != nil {
		return applicationDirectories{}, err
	}

	return directories, nil
}

func saveUploadedInput(inputDir, fileName string, data []byte) (string, error) {
	if err := os.MkdirAll(inputDir, 0o755); err != nil {
		return "", fmt.Errorf("không thể tạo thư mục đầu vào: %w", err)
	}
	path := filepath.Join(inputDir, safeStoredName(fileName))
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return "", fmt.Errorf("không thể lưu tệp đầu vào: %w", err)
	}
	return path, nil
}
