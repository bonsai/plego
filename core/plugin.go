package core

import "context"

type Item struct {
	ID    string // dedup key (e.g. filepath or hash)
	Title string
	Body  string
}

type Source interface {
	Name() string
	Items(ctx context.Context) ([]Item, error)
}

type Output interface {
	Name() string
	Publish(ctx context.Context, item Item) error
}

type StateStore interface {
	IsSent(id string) bool
	MarkSent(id string) error
	Close() error
}
