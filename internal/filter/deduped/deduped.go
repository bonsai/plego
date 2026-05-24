package deduped

import (
	"context"

	"github.com/dance/plego/core"
)

type Filter struct {
	store core.StateStore
}

func New(store core.StateStore) *Filter {
	return &Filter{store: store}
}

func (f *Filter) Name() string { return "Filter::Deduped" }

func (f *Filter) Filter(_ context.Context, entry *core.Entry) (*core.Entry, error) {
	if entry == nil {
		return nil, nil
	}

	guid := entry.GUID
	if guid == "" {
		guid = entry.URL
	}
	if guid == "" {
		return entry, nil
	}

	if f.store != nil && f.store.IsSent(guid) {
		return nil, nil
	}

	return entry, nil
}

var _ core.Filter = (*Filter)(nil)
