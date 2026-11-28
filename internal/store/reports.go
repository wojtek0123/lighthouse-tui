// Package store provides based operations to manage local files.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"uuid"

	"github.com/wojtek0123/lighthouse-tui/internal/report"
)

// ReportStore manages reports on local disk.
type ReportStore struct {
	dir string
}

// NewReportStore creates report store to manage report locally.
func NewReportStore(dir string) *ReportStore {
	return &ReportStore{dir: dir}
}

// Save stores a report on local disk in app dir.
func (s *ReportStore) Save(id uuid.UUID, data []byte) error {
	// data, err := json.Marshal(r)

	// if err != nil {
	// 	return fmt.Errorf("failed to marshal report %w", err)
	// }

	path := filepath.Join(s.dir, id.String()+".report.json")

	if err := os.WriteFile(path, data, writePermission); err != nil {
		return fmt.Errorf("failed to save report %w", err)
	}

	return nil
}

// Delete delets report from local disk. Make sure you deleted a report summary of this report.
func (s *ReportStore) Delete(id string) error {
	if err := os.Remove(filepath.Join(s.dir, id+".report.json")); err != nil {
		return fmt.Errorf("failed to delete report with %s ID %w", id, err)
	}
	return nil
}

// Load reads report from local disk based on ID.
func (s *ReportStore) Load(id uuid.UUID) (*report.Report, error) {
	data, err := os.ReadFile(filepath.Join(s.dir, id.String()+".report.json"))

	if err != nil {
		return nil, fmt.Errorf("read report %s: %w", id, err)
	}

	var r report.Report

	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("unmarshal report %s: %w", id, err)
	}

	return &r, nil
}
