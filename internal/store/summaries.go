package store

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"time"
)

type SummaryStore struct {
	dir string
}

type Score struct {
	Performance     int `json:"performance"`
	Accessibility   int `json:"accessibility"`
	BestPractices   int `json:"best_practices"`
	SEO             int `json:"seo"`
	AgenticBrowsing int `json:"agentic_browsing"`
}

type Summary struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	CreatedAt time.Time `json:"created_at"`
	Scores    Score     `json:"scores"`
}

type Summaries []Summary

func NewSummaryStore(dir string) *SummaryStore {
	return &SummaryStore{dir: dir}
}

func (s *SummaryStore) Save(summary *Summary) error {
	data, err := json.Marshal(summary)

	if err != nil {
		return fmt.Errorf("failed to marshal summary %w", err)
	}

	if err := os.WriteFile(s.dir, data, writePermission); err != nil {
		return fmt.Errorf("failed to save report summary %w", err)
	}

	return nil
}

func (s *SummaryStore) List() (Summaries, error) {
	entries, err := os.ReadDir(filepath.Join(s.dir, "*.summary.json"))

	if err != nil {
		return nil, err
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

func (s *SummaryStore) Delete(ID string) error {
	if err := os.Remove(filepath.Join(s.dir, ID+".summary.json")); err != nil {
		return err
	}
	return nil
}

func (s *SummaryStore) Load(path string) (Summary, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return Summary{}, err
	}

	var summary Summary
	if err := json.Unmarshal(data, &summary); err != nil {
		return Summary{}, err
	}

	return summary, nil
}

func (s Summary) Path() string {
	u, err := url.Parse(s.URL)

	if err != nil {
		return ""
	}

	return u.Path
}

func (s Summary) Hostname() string {
	u, err := url.Parse(s.URL)

	if err != nil {
		return ""
	}

	return u.Hostname()
}

func (ss Summaries) GroupByHostname() map[string]Summaries {
	summaryByHostname := map[string]Summaries{}

	for _, summary := range ss {
		summaryByHostname[summary.Hostname()] = append(summaryByHostname[summary.Hostname()], summary)
	}

	return summaryByHostname
}
