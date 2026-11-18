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

// CategoryID represents core web vitals category ID in lighthouse report.
type CategoryID string

const (
	// PerformanceCategoryID represents performance category in lighthouse report.
	PerformanceCategoryID CategoryID = "performance"

	// AccessibilityCategoryID represents accessibility category in lighthouse report.
	AccessibilityCategoryID CategoryID = "accessibility"

	// BestPractisesCategoryID represents best practises category in lighthouse report.
	BestPractisesCategoryID CategoryID = "best-practices"

	// SeoCategoryID represents seo category in lighthouse report.
	SeoCategoryID CategoryID = "seo"

	// AgenticBrowsingCategoryID represents agentic browsing category in lighthouse report.
	AgenticBrowsingCategoryID CategoryID = "agentic-browsing"
)

// Result represents report of lighthouse package and report name.
type Result struct {
	Report     *Report
	ReportName string
}

// ScoreDisplayMode specifies how a score or metric result is rendered in the UI.
type ScoreDisplayMode string

const (
	// MetricSavings displays the score formatted as calculated metric savings.
	MetricSavings ScoreDisplayMode = "metricSavings"

	// NonApplicable indicates that the score does not apply to the current context.
	NonApplicable ScoreDisplayMode = "notApplicable"

	// Informative presents the score as contextual information rather than a numerical grade.
	Informative ScoreDisplayMode = "informative"

	// Numeric displays the score as a quantitative numerical value.
	Numeric ScoreDisplayMode = "numeric"

	// Binary displays the score as a simple pass/fail or true/false outcome.
	Binary ScoreDisplayMode = "binary"

	// Manual indicates that the score requires manual evaluation or visual entry.
	Manual ScoreDisplayMode = "manual"

	// Error indicates that an error occurred while calculating or displaying the score.
	Error ScoreDisplayMode = "error"
)

// Report represents lighthouse reports in json format.
type Report struct {
	FinalURL          string `json:"finalUrl"`
	FetchTime         string `json:"fetchTime"`
	LighthouseVersion string `json:"lighthouseVersion"`
	Audits            map[string]struct {
		NumericValue float64 `json:"numericValue"`

		DisplayValue     string           `json:"displayValue"`
		Title            string           `json:"title"`
		Description      string           `json:"description"`
		Unit             string           `json:"numericUnit"`
		Score            *float64         `json:"score"`
		ScoreDisplayMode ScoreDisplayMode `json:"scoreDisplayMode"`
	} `json:"audits"`
	Categories map[CategoryID]Category
}

// Category represents specific category structure.
type Category struct {
	Title     string     `json:"title"`
	ID        CategoryID `json:"id"`
	Score     float64    `json:"score"`
	AuditRefs []AuditRefs
}

// AuditRefs represents available audits for category.
type AuditRefs struct {
	ID      string  `json:"id"`
	Weight  float64 `json:"weight"`
	Group   string  `json:"group"`
	Acronym string  `json:"acronym"`
}

// AuditGroup represents the category of audit logging or diagnostics grouping.
type AuditGroup string

const (
	// MetricsGroup identifies audit events related to performance and operational metrics.
	MetricsGroup AuditGroup = "metrics"

	// InsightGroup identifies audit events related to high-level system insights and analytics.
	InsightGroup AuditGroup = "insight"

	// DiagnosticsGroup identifies audit events used for technical troubleshooting and diagnostics.
	DiagnosticsGroup AuditGroup = "diagnostics"

	// HiddenGroup identifies audit events that are internal or hidden from standard user views.
	HiddenGroup AuditGroup = "hidden"
)

// RunLighthouse starts lighthouse command in headless mode to get report for specific website.
func RunLighthouse(ctx context.Context, url string) (*Result, error) {
	cmd := exec.CommandContext(
		ctx,
		"npx",
		"lighthouse",
		url,
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
		return nil, fmt.Errorf("failed to get lighthouse report for %s: %w", url, err)
	}

	filename := reportName()

	path, err := reportPath(url, filename)

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
