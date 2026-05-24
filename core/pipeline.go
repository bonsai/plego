package core

import (
	"context"
	"fmt"
	"log"
)

type Pipeline struct {
	Subscription Subscription
	Filters      []Filter
	Publishes    []Publish
	State        StateStore
}

func (p *Pipeline) Run(ctx context.Context) error {
	feeds, err := p.Subscription.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("subscription %s: %w", p.Subscription.Name(), err)
	}

	var entries []*Entry
	for _, feed := range feeds {
		log.Printf("[%s] feed: %s (%d entries)", p.Subscription.Name(), feed.Title, len(feed.Entries))
		entries = append(entries, feed.Entries...)
	}

	log.Printf("[%s] total %d entries", p.Subscription.Name(), len(entries))

	sent, skipped, dropped := 0, 0, 0
	for _, entry := range entries {
		if entry == nil {
			continue
		}

		guid := entry.GUID
		if guid == "" {
			guid = entry.URL
		}

		if p.State != nil && p.State.IsSent(guid) {
			skipped++
			continue
		}

		current := entry
		for _, filter := range p.Filters {
			var err error
			current, err = filter.Filter(ctx, current)
			if err != nil {
				log.Printf("[%s] filter error for %q: %v", filter.Name(), entry.Title, err)
				break
			}
			if current == nil {
				dropped++
				log.Printf("[%s] dropped: %s", filter.Name(), entry.Title)
				break
			}
		}
		if current == nil {
			continue
		}

		for _, pub := range p.Publishes {
			if err := pub.Publish(ctx, current); err != nil {
				log.Printf("[%s] publish error for %q: %v", pub.Name(), current.Title, err)
				continue
			}
			log.Printf("[%s] published: %s", pub.Name(), current.Title)
		}

		if p.State != nil {
			if err := p.State.MarkSent(guid); err != nil {
				log.Printf("state update error: %v", err)
			}
		}
		sent++
	}

	log.Printf("done: %d sent, %d skipped, %d dropped", sent, skipped, dropped)
	return nil
}
