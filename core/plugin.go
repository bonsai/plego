package core

import (
	"context"
	"time"
)

type Item struct {
	ID          string
	Title       string
	Body        string
	URL         string
	PublishedAt time.Time
	EventAt     *time.Time
	Location    string
}

type Source interface {
	Name() string
	Items(ctx context.Context) ([]Item, error)
}

// Filter keeps or drops items before they reach outputs.
type Filter interface {
	Name() string
	Keep(item Item) bool
}

type Output interface {
	Name() string
	Publish(ctx context.Context, item Item) error
}

// Flusher is implemented by outputs that batch items (digest email, iCal, JSON).
type Flusher interface {
	Flush(ctx context.Context) error
}

type StateStore interface {
	IsSent(id string) bool
	MarkSent(id string) error
	Close() error
}
