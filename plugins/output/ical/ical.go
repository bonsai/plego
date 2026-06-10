package ical

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dance/plego/core"
)

type Output struct {
	OutputPath string // e.g. "docs/calendar.ics"
	pending    []core.Item
}

func (o *Output) Name() string { return "ical" }

func (o *Output) Publish(_ context.Context, item core.Item) error {
	o.pending = append(o.pending, item)
	return nil
}

func (o *Output) Flush(_ context.Context) error {
	path := o.OutputPath
	if path == "" {
		path = "docs/calendar.ics"
	}

	// Load existing events if the file already exists.
	existing := loadExisting(path)

	// Merge: new events take precedence; existing events not in new batch are retained.
	merged := mergeEvents(existing, o.pending)

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	f.WriteString("BEGIN:VCALENDAR\r\n")
	f.WriteString("VERSION:2.0\r\n")
	f.WriteString("PRODID:-//Plego//Plego Digest//JA\r\n")
	f.WriteString("X-WR-CALNAME:飲食イベント新着\r\n")
	f.WriteString("X-WR-TIMEZONE:Asia/Tokyo\r\n")
	f.WriteString("CALSCALE:GREGORIAN\r\n")
	f.WriteString("METHOD:PUBLISH\r\n")

	for _, ev := range merged {
		f.WriteString(ev)
	}

	f.WriteString("END:VCALENDAR\r\n")

	log.Printf("[ical] wrote %d events to %s", len(merged), path)
	return nil
}

func itemToVEVENT(item core.Item) string {
	uid := eventUID(item.URL)
	now := time.Now().UTC().Format("20060102T150405Z")

	var sb strings.Builder
	sb.WriteString("BEGIN:VEVENT\r\n")
	sb.WriteString(fmt.Sprintf("UID:%s@plego\r\n", uid))
	sb.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", now))
	sb.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", icalEscape(item.Title)))

	if item.EventAt != nil {
		start := item.EventAt.UTC().Format("20060102T150405Z")
		end := item.EventAt.Add(time.Hour).UTC().Format("20060102T150405Z")
		sb.WriteString(fmt.Sprintf("DTSTART:%s\r\n", start))
		sb.WriteString(fmt.Sprintf("DTEND:%s\r\n", end))
	} else {
		date := item.PublishedAt.Format("20060102")
		sb.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", date))
		sb.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", date))
	}

	desc := item.URL
	if item.Body != "" {
		desc = item.Body + "\\n\\n" + item.URL
	}
	sb.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", icalEscape(desc)))
	sb.WriteString(fmt.Sprintf("URL:%s\r\n", item.URL))

	if item.Location != "" {
		sb.WriteString(fmt.Sprintf("LOCATION:%s\r\n", icalEscape(item.Location)))
	}
	sb.WriteString("END:VEVENT\r\n")
	return sb.String()
}

func eventUID(rawURL string) string {
	h := sha256.Sum256([]byte(rawURL))
	return fmt.Sprintf("%x", h[:16])
}

func icalEscape(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

// loadExisting reads UID → VEVENT blocks from an existing .ics file.
func loadExisting(path string) map[string]string {
	b, err := os.ReadFile(path)
	if err != nil {
		return map[string]string{}
	}
	events := map[string]string{}
	content := string(b)
	blocks := strings.Split(content, "BEGIN:VEVENT")
	for _, block := range blocks[1:] {
		end := strings.Index(block, "END:VEVENT")
		if end < 0 {
			continue
		}
		vevent := "BEGIN:VEVENT" + block[:end+len("END:VEVENT")] + "\r\n"
		// Extract UID
		for _, line := range strings.Split(vevent, "\n") {
			if strings.HasPrefix(line, "UID:") {
				uid := strings.TrimSpace(strings.TrimPrefix(line, "UID:"))
				events[uid] = vevent
				break
			}
		}
	}
	return events
}

// mergeEvents merges new items into existing events map, returns all VEVENT strings.
func mergeEvents(existing map[string]string, newItems []core.Item) []string {
	for _, item := range newItems {
		uid := eventUID(item.URL) + "@plego"
		existing[uid] = itemToVEVENT(item)
	}
	out := make([]string, 0, len(existing))
	for _, v := range existing {
		out = append(out, v)
	}
	return out
}

var _ core.Output = (*Output)(nil)
var _ core.Flusher = (*Output)(nil)
