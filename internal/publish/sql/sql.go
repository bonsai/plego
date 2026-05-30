package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dance/plego/core"
	"github.com/dance/plego/internal/yaqqle"
)

type Publish struct {
	OutDir string
}

func (p *Publish) Name() string {
	return "SQL"
}

func (p *Publish) Publish(ctx context.Context, entry *core.Entry) error {
	if entry == nil {
		return nil
	}

	var schema yaqqle.Schema
	if err := json.Unmarshal([]byte(entry.Body), &schema); err != nil {
		return fmt.Errorf("sql publish unmarshal: %w", err)
	}

	outDir := p.OutDir
	if outDir == "" {
		outDir = "."
	}

	for _, t := range schema.Tables {
		ddl := yaqqle.PostgresDDL(&t)
		outPath := filepath.Join(outDir, t.Name+".sql")
		if err := os.WriteFile(outPath, []byte(ddl), 0644); err != nil {
			return fmt.Errorf("sql publish write: %w", err)
		}
		log.Printf("[SQL] written: %s", outPath)
	}

	return nil
}
