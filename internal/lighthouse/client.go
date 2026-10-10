package lighthouse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

const appName string = "lighthouse-tui"

type Category string

const (
	Performance     Category = "performance"
	Accessibility   Category = "accessibility"
	BestPractises   Category = "best-practices"
	SEO             Category = "seo"
	AgenticBrowsing Category = "agentic-browsing"
)

type LighthouseResult struct {
	Report     LighthouseReport
	ReportName string
}

// check if lightyhouse version is supported
type LighthouseReport struct {
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
		Id    string  `json:"id"`
		Score float64 `json:"score"`
	}
}

func RunLighthouse(ctx context.Context, target string) (LighthouseResult, error) {

	cmd := exec.CommandContext(
		ctx,
		"npx",
		"lighthouse",
		target,
		"--quiet",
		"--output=json",
		"--chrome-flags=--headless=new",
	)

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
		return LighthouseResult{}, err
	}

	filename := reportName()

	path, err := reportPath(target, filename)

	if err != nil {
		return LighthouseResult{}, err
	}

	var report LighthouseReport
	err = json.Unmarshal(output, &report)

	err = os.WriteFile(path, output, 0o644)

	return LighthouseResult{
		Report:     report,
		ReportName: filename,
	}, err
}

func reportName() string {
	now := time.Now().Format(time.RFC3339)
	name := fmt.Sprintf("report-%s.json", now)

	return name
}

func reportPath(text string, fileName string) (string, error) {
	// now := time.Now().Format(time.RFC3339)
	// file := fmt.Sprintf("report-%s.json", now)
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	u, err := url.Parse(text)

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports", u.Host)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, fileName), nil
}

func GetReports(url string) ([]string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return make([]string, 0), err
	}

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports", url)

	entries, err := os.ReadDir(dir)

	if err != nil {
		return make([]string, 0), err
	}

	fileNames := []string{}

	for _, entry := range entries {
		if entry.IsDir() == false {
			fileNames = append(fileNames, entry.Name())
		}
	}

	return fileNames, err
}

func GetReport(hostname, reportName string) (LighthouseReport, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return LighthouseReport{}, err
	}

	base := filepath.Join(home, ".local", "share")
	path := filepath.Join(base, appName, "reports", hostname, reportName)

	output, err := os.ReadFile(path)

	var report LighthouseReport

	err = json.Unmarshal(output, &report)

	return report, err
}

func GetHostnames() ([]string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return make([]string, 0), err
	}

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports")

	entries, err := os.ReadDir(dir)

	if err != nil {
		return make([]string, 0), err
	}

	hostnames := []string{}

	for _, entry := range entries {
		if entry.IsDir() == true {
			hostnames = append(hostnames, entry.Name())
		}
	}

	return hostnames, err
}

func RemoveHostname(hostname string) error {
	home, err := os.UserHomeDir()

	if err != nil {
		return err
	}

	base := filepath.Join(home, ".local", "share")
	path := filepath.Join(base, appName, "reports", hostname)

	err = os.RemoveAll(path)

	return err
}
