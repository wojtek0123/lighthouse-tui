package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/wojtek0123/lighthouse-tui/internal/lighthouse"
)

type ReportStore struct {
	dir string
}

// NewReportStore creates report store to manage report locally.
func NewReportStore(dir string) *ReportStore {
	return &ReportStore{dir: dir}
}

// Save stores a report on local disk in app dir.
func (s *ReportStore) Save(r lighthouse.Report) error {

	data, err := json.Marshal(r)

	if err != nil {
		return fmt.Errorf("failed to marshal report %w", err)
	}

	hostname, err := Hostname(r)

	if err != nil {
		return err
	}

	id := newReportID()
	path := filepath.Join(s.dir, hostname, "reports", id)

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("failed to save report %w", err)
	}

	return nil
}

func (s *ReportStore) Delete(ID string) error {
	return nil
}

func (s *ReportStore) DeleteAll() error {
	return nil
}

func (s *ReportStore) List() {}

func (s *ReportStore) Load() {}

func newReportID() string {
	now := time.Now().Format(time.RFC3339)

	return fmt.Sprintf("report-%s.json", now)
}
