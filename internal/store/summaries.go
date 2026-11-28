package store

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"time"
	"uuid"
)

// SummaryStore represents store state.
type SummaryStore struct {
	dir string
}

// Score represents field in summary which holds basic score of lighthouse categories.
type Score struct {
	Performance     int `json:"performance"`
	Accessibility   int `json:"accessibility"`
	BestPractices   int `json:"best_practices"`
	SEO             int `json:"seo"`
	AgenticBrowsing int `json:"agentic_browsing"`
}

// Summary represents basic metadata about a lighthouse report.
type Summary struct {
	ID        uuid.UUID `json:"id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Scores    Score     `json:"scores"`
}

// Summaries represents slice of summaries.
type Summaries []Summary

// NewSummaryStore creates summary store instance.
func NewSummaryStore(dir string) *SummaryStore {
	return &SummaryStore{dir: dir}
}

// Save writes file with summary.json suffix on local disk.
func (s *SummaryStore) Save(summary *Summary) error {
	data, err := json.Marshal(summary)

	if err != nil {
		return fmt.Errorf("failed to marshal summary %w", err)
	}

	if err := os.WriteFile(filepath.Join(s.dir, summary.ID.String()+".summary.json"), data, writePermission); err != nil {
		return fmt.Errorf("failed to save report summary %w", err)
	}

	return nil
}

// List returns slice of summaries which are saved on local disk.
func (s *SummaryStore) List() (Summaries, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "*.summary.json"))

	if err != nil {
		return nil, fmt.Errorf("failed to read summary files at application directory %w", err)
	}

	if len(entries) == 0 {
		return Summaries{}, nil
	}

	summaries := make(Summaries, 0, len(entries))

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		summary, err := s.Load(filepath.Join(s.dir, entry.Name()))

		if err != nil {
			continue
		}

		summaries = append(summaries, summary)
	}

	return summaries, nil
}

// Delete deletes summary by id.
func (s *SummaryStore) Delete(id string) error {
	if err := os.Remove(filepath.Join(s.dir, id+".summary.json")); err != nil {
		return fmt.Errorf("failed to delete summary with ID %s %w", id, err)
	}
	return nil
}

// Load reads summary from a saved json file inside the store's base directory.
func (s *SummaryStore) Load(name string) (Summary, error) {
	root, err := os.OpenRoot(s.dir)
	if err != nil {
		return Summary{}, fmt.Errorf("failed to open store dir %s: %w", s.dir, err)
	}

	defer func() {
		if cerr := root.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close root: %w", cerr)
		}
	}()

	f, err := root.Open(name)
	if err != nil {
		return Summary{}, fmt.Errorf("failed to open file %s: %w", name, err)
	}

	defer func() {
		if cerr := f.Close(); cerr != nil && err == nil {
			err = fmt.Errorf("close file: %w", cerr)
		}
	}()

	var summary Summary
	if err := json.NewDecoder(f).Decode(&summary); err != nil {
		return Summary{}, fmt.Errorf("failed to decode summary: %w", err)
	}

	return summary, nil
}

// Path returns path from summary.
func (s Summary) Path() string {
	u, err := url.Parse(s.URL)

	if err != nil {
		return ""
	}

	return u.Path
}

// Hostname returns hostname of a summary url.
func (s Summary) Hostname() string {
	u, err := url.Parse(s.URL)

	if err != nil {
		return ""
	}

	return u.Hostname()
}

// GroupByHostname groups summaries by hostname.
func (ss Summaries) GroupByHostname() map[string]Summaries {
	summaryByHostname := map[string]Summaries{}

	for _, summary := range ss {
		summaryByHostname[summary.Hostname()] = append(summaryByHostname[summary.Hostname()], summary)
	}

	return summaryByHostname
}

// Sort sort summaries by created at field.
func (ss Summaries) Sort() {
	slices.SortFunc(ss, func(i, j Summary) int {
		return i.CreatedAt.Compare(j.CreatedAt)
	})
}
