package core

import (
	"context"
	"fmt"
	"log"
)

type Pipeline struct {
	Source  Source
	Outputs []Output
	State   StateStore
}

func (p *Pipeline) Run(ctx context.Context) error {
	items, err := p.Source.Items(ctx)
	if err != nil {
		return fmt.Errorf("source %s: %w", p.Source.Name(), err)
	}

	log.Printf("[%s] %d items found", p.Source.Name(), len(items))

	sent, skipped := 0, 0
	for _, item := range items {
		if p.State.IsSent(item.ID) {
			skipped++
			continue
		}

		for _, out := range p.Outputs {
			if err := out.Publish(ctx, item); err != nil {
				log.Printf("[%s] publish error for %q: %v", out.Name(), item.Title, err)
				continue
			}
			log.Printf("[%s] published: %s", out.Name(), item.Title)
		}

		if err := p.State.MarkSent(item.ID); err != nil {
			log.Printf("state update error: %v", err)
		}
		sent++
	}

	log.Printf("done: %d sent, %d skipped", sent, skipped)
	return nil
}
