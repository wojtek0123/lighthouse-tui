package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"uuid"

	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

const reportFileExtension = "json"

type ReportStore struct {
	dir string
}

// NewReportStore creates report store to manage report locally.
func NewReportStore(dir string) *ReportStore {
	return &ReportStore{dir: dir}
}

// Save stores a report on local disk in app dir.
func (s *ReportStore) Save(ID uuid.UUID, r lighthouse.Report) error {
	data, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("failed to marshal report %w", err)
	}

	path := filepath.Join(s.dir, ID.String()+".report.json")

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to save report %w", err)
	}

	return nil
}

func (s *ReportStore) Delete(ID string) error {
	if err := os.Remove(path.Join(s.dir, ID+"."+reportFileExtension)); err != nil {
		return err
	}
	return nil
}

func (s *ReportStore) DeleteAll() error {
	return nil
}

func (s *ReportStore) List() (map[string][]lighthouse.Report, error) {
	hostnames, err := getHostnames(s.dir)

	if err != nil {
		return nil, err
	}

	reportsByHostname := map[string][]lighthouse.Report{}

	for _, hostname := range hostnames {
		reports, err := getReports(filepath.Join(s.dir, hostname, "reports"))

		if err != nil {
			return nil, err
		}

		reportsByHostname[hostname] = append(reportsByHostname[hostname], reports...)

	}

	return reportsByHostname, nil
}

func getHostnames(path string) ([]string, error) {
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			hosts = append(hosts, e.Name())
		}
	}

	return hosts, nil
}

func getReports(path string) ([]lighthouse.Report, error) {
	entries, err := os.ReadDir(path)

	if err != nil {
		return nil, err
	}

	reports := make([]lighthouse.Report, 0, len(entries))

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		data, err := os.ReadFile(path)

		if err != nil {
			return nil, err
		}

		var report lighthouse.Report

		if err := json.Unmarshal(data, &report); err != nil {
			return nil, err
		}

		reports = append(reports, report)
	}

	return reports, nil
}
