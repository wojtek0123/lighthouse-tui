package store

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
)

const app = "lighthouse-tui"

const (
	writePermission = 0o600
	makePermission  = 0o750
)

// DefaultDir returns default app directory to store data.
func DefaultDir() (string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("failed to get home dir %w", err)
	}

	return filepath.Join(home, ".local", "share", app), nil
}

// Hostname returns hostname of a fetched url report.
func Hostname(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)

	if err != nil {
		return "", fmt.Errorf("failed to parse report url %w", err)
	}

	return u.Hostname(), nil
}

func CreateDir(path string) error {
	if err := os.MkdirAll(path, makePermission); err != nil {
		return fmt.Errorf("failed to create directory at %s for hostname: %w", path, err)
	}

	return nil
}
