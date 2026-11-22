package store

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
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
func Hostname(r lighthouse.Report) (string, error) {
	u, err := url.Parse(r.FinalURL)

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
