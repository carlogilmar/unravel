package config

import (
	"os"
	"path/filepath"
)

func DataDir() (string, error) {
	if x := os.Getenv("XDG_DATA_HOME"); x != "" {
		return ensure(filepath.Join(x, "unravel"))
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return ensure(filepath.Join(home, ".local", "share", "unravel"))
}

func DBPath() (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "unravel.db"), nil
}

func ensure(dir string) (string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}
