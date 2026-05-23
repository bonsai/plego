package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dance/plego/core"
)

type Source struct {
	Path       string
	Extensions []string // e.g. [".md", ".txt"]
}

func (s *Source) Name() string { return "filesystem" }

func (s *Source) Items(_ context.Context) ([]core.Item, error) {
	exts := s.Extensions
	if len(exts) == 0 {
		exts = []string{".md", ".txt"}
	}

	var items []core.Item
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

		items = append(items, core.Item{
			ID:    path,
			Title: title(path, string(body)),
			Body:  string(body),
		})
		return nil
	})

	return items, err
}

func title(path, body string) string {
	// first non-empty line, stripped of markdown heading markers
	for _, line := range strings.SplitN(body, "\n", 20) {
		line = strings.TrimSpace(strings.TrimLeft(line, "# "))
		if line != "" {
			return line
		}
	}
	return filepath.Base(path)
}
