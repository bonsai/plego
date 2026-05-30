package yaml

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dance/plego/core"
)

type Subscription struct {
	Path string
}

func (s *Subscription) Name() string {
	return "YAML"
}

func (s *Subscription) Fetch(ctx context.Context) ([]*core.Feed, error) {
	fi, err := os.Stat(s.Path)
	if err != nil {
		return nil, fmt.Errorf("yaml subscription: %w", err)
	}

	var files []string
	if fi.IsDir() {
		entries, err := os.ReadDir(s.Path)
		if err != nil {
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".yaml" {
				files = append(files, filepath.Join(s.Path, e.Name()))
			}
		}
	} else {
		files = append(files, s.Path)
	}

	if len(files) == 0 {
		log.Printf("[YAML] no .yaml files found at %s", s.Path)
		return nil, nil
	}

	var feeds []*core.Feed
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Printf("[YAML] skip %s: %v", f, err)
			continue
		}

		feed := &core.Feed{
			GUID:    f,
			Title:   filepath.Base(f),
			URL:     f,
			Content: string(b),
			Entries: []*core.Entry{
				{
					GUID:  f,
					Title: filepath.Base(f),
					URL:   f,
					Body:  string(b),
				},
			},
		}
		feeds = append(feeds, feed)
	}

	return feeds, nil
}
