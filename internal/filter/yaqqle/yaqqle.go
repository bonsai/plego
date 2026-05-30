package yaqqlefilter

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dance/plego/core"
	"github.com/dance/plego/internal/yaqqle"
)

type Filter struct{}

func New() *Filter {
	return &Filter{}
}

func (f *Filter) Name() string {
	return "YaQQle"
}

func (f *Filter) Filter(ctx context.Context, entry *core.Entry) (*core.Entry, error) {
	if entry == nil {
		return nil, nil
	}

	schema, err := yaqqle.ParseFile(entry.URL)
	if err != nil {
		return nil, fmt.Errorf("yaqqle filter: %w", err)
	}

	b, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("yaqqle filter marshal: %w", err)
	}

	entry.Body = string(b)
	return entry, nil
}
