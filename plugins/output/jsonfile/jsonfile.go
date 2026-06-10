package jsonfile

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/dance/plego/core"
)

type record struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	URL         string     `json:"url"`
	PublishedAt time.Time  `json:"published_at"`
	EventAt     *time.Time `json:"event_at"`
	Location    string     `json:"location"`
}

type Output struct {
	OutputPath string // e.g. "docs/feed.json"
	pending    []core.Item
}

func (o *Output) Name() string { return "json" }

func (o *Output) Publish(_ context.Context, item core.Item) error {
	o.pending = append(o.pending, item)
	return nil
}

func (o *Output) Flush(_ context.Context) error {
	path := o.OutputPath
	if path == "" {
		path = "docs/feed.json"
	}

	// Merge with existing records so historical items are preserved.
	existing := loadExisting(path)
	byID := make(map[string]record, len(existing))
	for _, r := range existing {
		byID[r.ID] = r
	}
	for _, item := range o.pending {
		byID[item.ID] = record{
			ID:          item.ID,
			Title:       item.Title,
			Body:        item.Body,
			URL:         item.URL,
			PublishedAt: item.PublishedAt,
			EventAt:     item.EventAt,
			Location:    item.Location,
		}
	}

	merged := make([]record, 0, len(byID))
	for _, r := range byID {
		merged = append(merged, r)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(merged); err != nil {
		return err
	}
	log.Printf("[json] wrote %d records to %s", len(merged), path)
	return nil
}

func loadExisting(path string) []record {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []record
	_ = json.Unmarshal(b, &out)
	return out
}

var _ core.Output = (*Output)(nil)
var _ core.Flusher = (*Output)(nil)
