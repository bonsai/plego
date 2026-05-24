package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/dance/plego/config"
	"github.com/dance/plego/core"
	"github.com/dance/plego/internal/filter/deduped"
	"github.com/dance/plego/internal/filter/html2text"
	"github.com/dance/plego/internal/publish/googledrive"
	"github.com/dance/plego/internal/subscription/googledrivesub"
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

	stateDB := cfg.Global.StateDB
	if stateDB == "" {
		stateDB = filepath.Join(os.Getenv("USERPROFILE"), ".plego", "state.db")
	}

	store, err := state.Open(stateDB)
	if err != nil {
		log.Fatalf("state: %v", err)
	}
	defer store.Close()

	var subscription core.Subscription
	var filters []core.Filter
	var publishes []core.Publish

	for _, pc := range cfg.Plugins {
		plugin, err := buildPlugin(pc, store)
		if err != nil {
			log.Fatalf("plugin %s: %v", pc.Module, err)
		}
		switch p := plugin.(type) {
		case core.Subscription:
			if subscription != nil {
				log.Fatalf("multiple subscriptions not supported")
			}
			subscription = p
		case core.Filter:
			filters = append(filters, p)
		case core.Publish:
			publishes = append(publishes, p)
		default:
			log.Fatalf("unknown plugin type: %s", pc.Module)
		}
	}

	if subscription == nil {
		log.Fatalf("no subscription plugin configured")
	}

	ctx := context.Background()

	log.Println(">> initializing authentication...")
	if auth, ok := subscription.(core.Authorizer); ok {
		if err := auth.InitAuth(ctx); err != nil {
			log.Fatalf("subscription auth: %v", err)
		}
	}
	for _, f := range filters {
		if auth, ok := f.(core.Authorizer); ok {
			if err := auth.InitAuth(ctx); err != nil {
				log.Fatalf("filter auth: %v", err)
			}
		}
	}
	for _, p := range publishes {
		if auth, ok := p.(core.Authorizer); ok {
			if err := auth.InitAuth(ctx); err != nil {
				log.Fatalf("publish auth: %v", err)
			}
		}
	}

	pipeline := &core.Pipeline{
		Subscription: subscription,
		Filters:      filters,
		Publishes:    publishes,
		State:        store,
	}

	log.Println(">> starting pipeline...")
	if err := pipeline.Run(ctx); err != nil {
		log.Fatalf("pipeline: %v", err)
	}
}

func buildPlugin(pc config.PluginConfig, store core.StateStore) (interface{}, error) {
	cfg := pc.Config
	if cfg == nil {
		cfg = map[string]interface{}{}
	}

	switch pc.Module {

	// Subscriptions
	case "Subscription::Filesystem":
		return &filesystem.Subscription{
			Path:       stringValue(cfg, "path"),
			Extensions: stringSlice(cfg, "extensions"),
		}, nil

	case "Subscription::GoogleDrive":
		return googledrivesub.New(
			stringValue(cfg, "fileId"),
			stringValue(cfg, "credentials"),
		)

	// Filters
	case "Filter::HTML2Text":
		return html2text.New(), nil

	case "Filter::Deduped":
		return deduped.New(store), nil

	// Publishes
	case "Publish::Gmail":
		token := stringValue(cfg, "token")
		if token == "" {
			token = filepath.Join(os.Getenv("USERPROFILE"), ".plego", "gmail-token.json")
		}
		return &gmail.Publish{
			To:              stringValue(cfg, "to"),
			CredentialsFile: stringValue(cfg, "credentials"),
			TokenFile:       token,
		}, nil

	case "Publish::GoogleDrive":
		return googledrive.New(
			stringValue(cfg, "folderId"),
			stringValue(cfg, "fileNamePrefix"),
			stringValue(cfg, "timestampFormat"),
			stringValue(cfg, "extension"),
			stringValue(cfg, "credentials"),
		)

	default:
		return nil, fmt.Errorf("unknown module: %s", pc.Module)
	}
}

func stringValue(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

func intValue(m map[string]interface{}, key string) int {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case float64:
		return int(n)
	}
	return 0
}

func stringSlice(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok {
		return nil
	}
	raw, ok := v.([]interface{})
	if !ok {
		return nil
	}
	var out []string
	for _, item := range raw {
		if s, ok := item.(string); ok {
			out = append(out, s)
		}
	}
	return out
}

// unused import guard
var _ = strings.TrimSpace
