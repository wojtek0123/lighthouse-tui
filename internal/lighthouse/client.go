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

type Category string

const (
	Performance     Category = "performance"
	Accessibility   Category = "accessibility"
	BestPractises   Category = "best-practices"
	SEO             Category = "seo"
	AgenticBrowsing Category = "agentic-browsing"
)

// check if lightyhouse version is supported
type LighthouseResult struct {
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
	path, err := reportPath(target)

	if err != nil {
		return LighthouseResult{}, err
	}

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

	var result LighthouseResult
	err = json.Unmarshal(output, &result)

	os.WriteFile(path, output, 0o644)
	return result, err
}

func reportPath(text string) (string, error) {
	now := time.Now().Format(time.RFC3339)
	file := fmt.Sprintf("report-%s.json", now)
	home, err := os.UserHomeDir()

	if err != nil {
		return "", err
	}

	u, err := url.Parse(text)

	appName := "lighthouse-tui"

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports", u.Host)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	return filepath.Join(dir, file), nil
}

func GetReports(url string) ([]string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return make([]string, 0), err
	}

	appName := "lighthouse-tui"

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

func GetUrls() ([]string, error) {
	home, err := os.UserHomeDir()

	if err != nil {
		return make([]string, 0), err
	}

	appName := "lighthouse-tui"

	base := filepath.Join(home, ".local", "share")
	dir := filepath.Join(base, appName, "reports")

	entries, err := os.ReadDir(dir)

	if err != nil {
		return make([]string, 0), err
	}

	urls := []string{}

	for _, entry := range entries {
		if entry.IsDir() == true {
			urls = append(urls, entry.Name())
		}
	}

	return urls, err
}
