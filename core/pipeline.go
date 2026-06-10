package core

import (
	"context"
	"fmt"
	"log"
)

type Pipeline struct {
	Source  Source
	Filters []Filter
	Outputs []Output
	State   StateStore
}

func (p *Pipeline) Run(ctx context.Context) error {
	items, err := p.Source.Items(ctx)
	if err != nil {
		return fmt.Errorf("source %s: %w", p.Source.Name(), err)
	}
	log.Printf("[%s] %d items fetched", p.Source.Name(), len(items))

	// Apply filters.
	var filtered []Item
	for _, item := range items {
		pass := true
		for _, f := range p.Filters {
			if !f.Keep(item) {
				log.Printf("[%s] drop: %s", f.Name(), item.Title)
				pass = false
				break
			}
		}
		if pass {
			filtered = append(filtered, item)
		}
	}
	log.Printf("%d items after filters", len(filtered))

	sent, skipped := 0, 0
	for _, item := range filtered {
		if p.State.IsSent(item.ID) {
			skipped++
			continue
		}
		for _, out := range p.Outputs {
			if err := out.Publish(ctx, item); err != nil {
				log.Printf("[%s] publish error for %q: %v", out.Name(), item.Title, err)
			}
		}
		if err := p.State.MarkSent(item.ID); err != nil {
			log.Printf("state update error: %v", err)
		}
		sent++
	}

	// Flush batch outputs.
	for _, out := range p.Outputs {
		if f, ok := out.(Flusher); ok {
			if err := f.Flush(ctx); err != nil {
				log.Printf("[%s] flush error: %v", out.Name(), err)
			}
		}
	}

	log.Printf("done: %d sent, %d skipped (dedup)", sent, skipped)
	return nil
}
