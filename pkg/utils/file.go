package utils

import (
	"os"
	"path/filepath"
	"strings"
)


func ExpandTilde(path string) string {
	homeDir, err := os.UserHomeDir()

	if err != nil {
		panic("could not determine home directory")
	}

	if path == "~" {
		return homeDir
	}

	if strings.HasPrefix(path, "~/") {
		return filepath.Join(
			homeDir,
			path[2:],
		)
	}

	//TODO: Handle ~username

	return path
}
