package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dance/plego/core"
)

type Subscription struct {
	Path       string
	Extensions []string
}

func (s *Subscription) Name() string { return "Subscription::Filesystem" }

func (s *Subscription) Fetch(_ context.Context) ([]*core.Feed, error) {
	exts := s.Extensions
	if len(exts) == 0 {
		exts = []string{".md", ".txt"}
	}

	var entries []*core.Entry
	err := filepath.WalkDir(s.Path, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		if !slices.Contains(exts, strings.ToLower(filepath.Ext(path))) {
			return nil
		}

		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}

		entries = append(entries, &core.Entry{
			GUID:  path,
			Title: title(path, string(body)),
			Body:  string(body),
			URL:   path,
		})
		return nil
	})

	if err != nil {
		return nil, err
	}

	return []*core.Feed{{
		GUID:    "filesystem:" + s.Path,
		Title:   "Filesystem: " + s.Path,
		Entries: entries,
	}}, nil
}

func title(path, body string) string {
	for _, line := range strings.SplitN(body, "\n", 20) {
		line = strings.TrimSpace(strings.TrimLeft(line, "# "))
		if line != "" {
			return line
		}
	}
	return filepath.Base(path)
}

var _ core.Subscription = (*Subscription)(nil)
