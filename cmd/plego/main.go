package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/dance/plego/config"
	"github.com/dance/plego/core"
	"github.com/dance/plego/plugins/output/gmail"
	"github.com/dance/plego/plugins/source/filesystem"
	"github.com/dance/plego/state"
)

func main() {
	cfgPath := flag.String("config", "plego.yaml", "config file")
	flag.Parse()

	cfg, err := config.Load(*cfgPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	stateDB := cfg.Pipeline.StateDB
	if stateDB == "" {
		stateDB = filepath.Join(os.Getenv("USERPROFILE"), ".plego", "state.db")
	}

	store, err := state.Open(stateDB)
	if err != nil {
		log.Fatalf("state: %v", err)
	}
	defer store.Close()

	src, err := buildSource(cfg.Pipeline.Source)
	if err != nil {
		log.Fatalf("source: %v", err)
	}

	var outputs []core.Output
	for _, outCfg := range cfg.Pipeline.Outputs {
		out, err := buildOutput(outCfg)
		if err != nil {
			log.Fatalf("output: %v", err)
		}
		outputs = append(outputs, out)
	}

	ctx := context.Background()

	// OAUTH: Initialize authentication for all plugins that need it
	log.Println(">> initializing authentication...")
	if src != nil {
		if auth, ok := src.(core.Authorizer); ok {
			if err := auth.InitAuth(ctx); err != nil {
				log.Fatalf("source auth: %v", err)
			}
		}
	}
	for _, out := range outputs {
		if auth, ok := out.(core.Authorizer); ok {
			if err := auth.InitAuth(ctx); err != nil {
				log.Fatalf("output auth: %v", err)
			}
		}
	}

	pipeline := &core.Pipeline{
		Source:  src,
		Outputs: outputs,
		State:   store,
	}

	// DO APP: Run the pipeline
	log.Println(">> starting pipeline...")
	if err := pipeline.Run(ctx); err != nil {
		log.Fatalf("pipeline: %v", err)
	}
}

func buildSource(cfg config.SourceConfig) (core.Source, error) {
	switch cfg.Module {
	case "filesystem":
		return &filesystem.Source{
			Path:       cfg.Path,
			Extensions: cfg.Extensions,
		}, nil
	default:
		log.Fatalf("unknown source module: %s", cfg.Module)
		return nil, nil
	}
}

func buildOutput(cfg config.OutputConfig) (core.Output, error) {
	switch cfg.Module {
	case "gmail":
		token := cfg.Token
		if token == "" {
			token = filepath.Join(os.Getenv("USERPROFILE"), ".plego", "gmail-token.json")
		}
		return &gmail.Output{
			To:              cfg.To,
			CredentialsFile: cfg.Credentials,
			TokenFile:       token,
		}, nil
	default:
		log.Fatalf("unknown output module: %s", cfg.Module)
		return nil, nil
	}
}
