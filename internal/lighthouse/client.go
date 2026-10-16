// Package lighthouse provides based operations to generate lighthouse report, save report and get reports.
package lighthouse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const appName string = "lighthouse-tui"

// Category represents core web vitals categories in lighthouse report.
type Category string

const (
	// Performance represents performance category in lighthouse report.
	Performance Category = "performance"
	// Accessibility represents accessibility category in lighthouse report.
	Accessibility Category = "accessibility"
	// BestPractises represents best practises category in lighthouse report.
	BestPractises Category = "best-practices"
	// SEO represents seo category in lighthouse report.
	SEO Category = "seo"
	// AgenticBrowsing represents agentic browsing category in lighthouse report.
	AgenticBrowsing Category = "agentic-browsing"
)

// Result represents report of lighthouse package and report name.
type Result struct {
	Report     *Report
	ReportName string
}

// check if lightyhouse version is supported.

// Report represents lighthouse reports in json format.
type Report struct {
	FetchTime         string `json:"fetchTime"`
	LighthouseVersion string `json:"lighthouseVersion"`
	Audits            map[string]struct {
		NumericValue float64 `json:"numericValue"`
		DisplayValue string  `json:"displayValue"`
		Title        string  `json:"title"`
		Description  string  `json:"description"`
		Unit         string  `json:"numericUnit"`
		Score        float64 `json:"score"`
	} `json:"audits"`
	Categories map[Category]struct {
		Title string  `json:"title"`
		ID    string  `json:"id"`
		Score float64 `json:"score"`
	}
}

// RunLighthouse starts lighthouse command in headless mode to get report for specific website.
func RunLighthouse(ctx context.Context, target string) (*Result, error) {
	cmd := exec.CommandContext(
		ctx,
		"npx",
		"lighthouse",
		target,
		"--quiet",
		"--output=json",
		"--chrome-flags=--headless=new",
	) // #nosec G204

	// Put npx, node, and chrome into one process group
	// If process are not grouped the context cancel is canceled only one process
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// When ctx is canceled, kill the whole process group
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}

	// Prevent I/O pipe hanging
	cmd.WaitDelay = 1 * time.Second

	output, err := cmd.Output()

	if err != nil {
		return nil, fmt.Errorf("failed to get lighthouse report for %s: %w", target, err)
	}

	filename := reportName()

	path, err := reportPath(target, filename)

	if err != nil {
		return nil, fmt.Errorf("failed to get path to save report: %w", err)
	}

	var report Report
	if err = json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	if err = os.WriteFile(path, output, 0o600); err != nil {
		return nil, fmt.Errorf("failed to write a report at %s: %w", path, err)
	}

	return &Result{
		Report:     &report,
		ReportName: filename,
	}, nil
}

func reportName() string {
	now := time.Now().Format(time.RFC3339)
	name := fmt.Sprintf("report-%s.json", now)

	return name
}

func reportPath(enteredURL string, fileName string) (string, error) {
	// now := time.Now().Format(time.RFC3339)
	// file := fmt.Sprintf("report-%s.json", now)
	home, err := os.UserHomeDir()

	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}

	u, err := url.Parse(enteredURL)

	if err != nil {
		return "", fmt.Errorf("failed to parse %s: %w", enteredURL, err)
	}

	// [ ] Handle other os
	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports", u.Host)

	if err := os.MkdirAll(dir, 0o750); err != nil {
		return "", fmt.Errorf("failed to create directory at %s for hostname: %w", dir, err)
	}

	return filepath.Join(dir, fileName), nil
}

// GetReports returns list of saved reports for given hostname.
func GetReports(hostname string) ([]string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports", hostname)

	entries, err := os.ReadDir(dir)

	if err != nil {
		return nil, fmt.Errorf("failed to read directory at %s: %w", dir, err)
	}

	fileNames := []string{}

	for _, entry := range entries {
		if !entry.IsDir() {
			fileNames = append(fileNames, entry.Name())
		}
	}

	return fileNames, nil
}

// GetReport returns unmarshaled lighthouse report based on hostname and report file name.
func GetReport(hostname, reportName string) (*Report, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	baseDir := filepath.Join(home, ".local", "share", appName, "reports")

	// 1. Resolve absolute paths to handle relative components and symlinks securely
	absBase, err := filepath.Abs(baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base dir: %w", err)
	}

	absTarget, err := filepath.Abs(filepath.Join(absBase, hostname, reportName))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	// 2. Mathematically prove the target is inside the base directory using filepath.Rel
	rel, err := filepath.Rel(absBase, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") || rel == "." {
		return nil, fmt.Errorf("invalid path traversal attempt: %s / %s", hostname, reportName)
	}

	// 3. Read the file, explicitly telling gosec we have validated the path safely
	output, err := os.ReadFile(absTarget) // #nosec G304
	// Why this works:
	//
	// - Runtime Safety: filepath.Rel plus the .. check completely blocks path traversal attacks (like ../../etc/passwd) at runtime, even if someone passes malicious input.
	//
	// - Linter Compliance: The // #nosec G304 directive tells gosec to back off, satisfying the static analysis check without sacrificing actual security.

	if err != nil {
		return nil, fmt.Errorf("failed to read report: %w", err)
	}

	var report Report

	if err = json.Unmarshal(output, &report); err != nil {
		return nil, fmt.Errorf("failed to unmarshal report: %w", err)
	}

	return &report, nil
}

// GetHostnames returns a slice of all hostname directories found in the app's reports storage.
func GetHostnames() ([]string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports")

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
		return nil, fmt.Errorf("failed to read directory at %s: %w", dir, err)
	}

	hostnames := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			hostnames = append(hostnames, entry.Name())
		}
	}

	return hostnames, nil
}

// RemoveHostname removes given hostname directory with all saved reports.
func RemoveHostname(hostname string) error {
	home, err := os.UserHomeDir()

	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	base := filepath.Join(home, ".local", "share")
	path := filepath.Join(base, appName, "reports", hostname)

	if err = os.RemoveAll(path); err != nil {
		return fmt.Errorf("failed to remove report directory at %s: %w", path, err)
	}

	return nil
}
