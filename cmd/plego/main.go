package main

import (
	"context"
	"flag"
	"log"
	"os"
	"path/filepath"

	"github.com/dance/plego/config"
	"github.com/dance/plego/core"
	"github.com/dance/plego/plugins/filter/yotei"
	"github.com/dance/plego/plugins/output/gmail"
	"github.com/dance/plego/plugins/output/ical"
	"github.com/dance/plego/plugins/output/jsonfile"
	"github.com/dance/plego/plugins/output/smtp"
	"github.com/dance/plego/plugins/source/filesystem"
	"github.com/dance/plego/plugins/source/prtimes"
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
		home := os.Getenv("HOME")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		stateDB = filepath.Join(home, ".plego", "state.db")
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

	var filters []core.Filter
	for _, fc := range cfg.Pipeline.Filters {
		f, err := buildFilter(fc)
		if err != nil {
			log.Fatalf("filter: %v", err)
		}
		filters = append(filters, f)
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
		Filters: filters,
		Outputs: outputs,
		State:   store,
	}

	log.Println(">> starting pipeline...")
	if err := pipeline.Run(ctx); err != nil {
		log.Fatalf("pipeline: %v", err)
	}
}

func buildSource(cfg config.SourceConfig) (core.Source, error) {
	switch cfg.Module {
	case "prtimes":
		return &prtimes.Source{
			Keywords:   cfg.Keywords,
			Industries: cfg.Industries,
		}, nil
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

func buildFilter(cfg config.FilterConfig) (core.Filter, error) {
	switch cfg.Module {
	case "yotei":
		return &yotei.Filter{}, nil
	default:
		log.Fatalf("unknown filter module: %s", cfg.Module)
		return nil, nil
	}
}

func buildOutput(cfg config.OutputConfig) (core.Output, error) {
	switch cfg.Module {
	case "smtp":
		return &smtp.Output{
			From:     cfg.From,
			Password: cfg.Password,
			To:       cfg.To,
			BCC:      cfg.BCC,
			Subject:  cfg.Subject,
		}, nil
	case "ical":
		return &ical.Output{
			OutputPath: cfg.OutputPath,
		}, nil
	case "json":
		return &jsonfile.Output{
			OutputPath: cfg.OutputPath,
		}, nil
	case "gmail":
		if len(cfg.To) == 0 {
			log.Fatalf("gmail output: 'to' must have at least one address")
		}
		tokenPath := cfg.Token
		if tokenPath == "" {
			home := os.Getenv("HOME")
			if home == "" {
				home = os.Getenv("USERPROFILE")
			}
			tokenPath = filepath.Join(home, ".plego", "gmail-token.json")
		}
		return &gmail.Output{
			To:              cfg.To[0],
			CredentialsFile: cfg.Credentials,
			TokenFile:       tokenPath,
		}, nil
	default:
		log.Fatalf("unknown output module: %s", cfg.Module)
		return nil, nil
	}
}
