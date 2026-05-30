package core

import "context"

type Feed struct {
	GUID    string
	Title   string
	URL     string
	Content string
	Entries []*Entry
}

type Entry struct {
	GUID  string
	Title string
	URL   string
	Body  string
}

type Subscription interface {
	Name() string
	Fetch(ctx context.Context) ([]*Feed, error)
}

type Filter interface {
	Name() string
	Filter(ctx context.Context, entry *Entry) (*Entry, error)
}

type Publish interface {
	Name() string
	Publish(ctx context.Context, entry *Entry) error
}

type StateStore interface {
	IsSent(id string) bool
	MarkSent(id string) error
	Close() error
}
