package json

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
	OutDir   string
	Schema   bool
}

func (p *Publish) Name() string {
	return "JSON"
}

func (p *Publish) Publish(ctx context.Context, entry *core.Entry) error {
	if entry == nil {
		return nil
	}

	var schema yaqqle.Schema
	if err := json.Unmarshal([]byte(entry.Body), &schema); err != nil {
		return fmt.Errorf("json publish unmarshal: %w", err)
	}

	outDir := p.OutDir
	if outDir == "" {
		outDir = "."
	}

	baseName := entry.Title
	if ext := filepath.Ext(baseName); ext != "" {
		baseName = baseName[:len(baseName)-len(ext)]
	}

	metaPath := filepath.Join(outDir, baseName+".json")
	b, err := json.MarshalIndent(&schema, "", "  ")
	if err != nil {
		return fmt.Errorf("json publish marshal: %w", err)
	}
	if err := os.WriteFile(metaPath, b, 0644); err != nil {
		return err
	}
	log.Printf("[JSON] written: %s", metaPath)

	if p.Schema {
		jsPath := filepath.Join(outDir, baseName+".schema.json")
		if err := yaqqle.GenerateJSONSchema(&schema, jsPath); err != nil {
			return err
		}
		log.Printf("[JSON] written: %s", jsPath)
	}

	return nil
}
