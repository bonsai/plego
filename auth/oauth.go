package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthHandler manages OAuth authentication with automatic fallback.
type OAuthHandler struct {
	Config          *oauth2.Config
	TokenFile       string
	CredentialsFile string
	Scopes          []string
	// FallbackEnvVar is the environment variable to check for a pre-stored token
	// If empty, defaults to {PLUGIN_NAME}_TOKEN
	FallbackEnvVar string
	// PluginName is used for default env var names and logging
	PluginName string
}

// Token loads a token, trying (in order):
// 1. Token file (persisted from previous run)
// 2. Environment variable (PLUGIN_TOKEN or custom)
// 3. Interactive OAuth flow
func (h *OAuthHandler) Token(ctx context.Context) (*oauth2.Token, error) {
	// Try: Load from file
	if tok, err := h.loadTokenFile(); err == nil {
		log.Printf("[%s] using existing token from %s", h.PluginName, h.TokenFile)
		return tok, nil
	}

	// Try: Load from environment variable (fallback)
	envVar := h.FallbackEnvVar
	if envVar == "" {
		envVar = strings.ToUpper(h.PluginName) + "_TOKEN"
	}
	if tok, err := h.loadTokenEnv(envVar); err == nil {
		log.Printf("[%s] using token from env var %s", h.PluginName, envVar)
		return tok, nil
	}

	// Final: Interactive OAuth flow
	log.Printf("[%s] no token found, starting OAuth flow...", h.PluginName)
	return h.interactiveAuth(ctx)
}

// loadTokenFile tries to load from token file
func (h *OAuthHandler) loadTokenFile() (*oauth2.Token, error) {
	if h.TokenFile == "" {
		return nil, fmt.Errorf("no token file configured")
	}
	f, err := os.Open(h.TokenFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var tok oauth2.Token
	if err := json.NewDecoder(f).Decode(&tok); err != nil {
		return nil, err
	}
	return &tok, nil
}

// loadTokenEnv tries to load from environment variable (JSON or raw token string)
func (h *OAuthHandler) loadTokenEnv(envVar string) (*oauth2.Token, error) {
	val := os.Getenv(envVar)
	if val == "" {
		return nil, fmt.Errorf("env var not set: %s", envVar)
	}

	// Try: Parse as JSON (full token)
	var tok oauth2.Token
	if err := json.Unmarshal([]byte(val), &tok); err == nil && tok.AccessToken != "" {
		return &tok, nil
	}

	// Try: Treat as raw access token
	if val != "" {
		return &oauth2.Token{
			AccessToken: val,
			TokenType:   "Bearer",
		}, nil
	}

	return nil, fmt.Errorf("invalid token in %s", envVar)
}

// interactiveAuth runs the OAuth flow with user interaction
func (h *OAuthHandler) interactiveAuth(ctx context.Context) (*oauth2.Token, error) {
	authURL := h.Config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("\n[%s] opening browser for authentication...\n", h.PluginName)
	fmt.Printf("If browser doesn't open, visit:\n  %s\n\n", authURL)
	fmt.Printf("Paste authorization code: ")

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return nil, fmt.Errorf("read auth code: %w", err)
	}
	code = strings.TrimSpace(code)

	tok, err := h.Config.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("exchange token: %w", err)
	}

	// Save for next time
	if err := h.saveTokenFile(tok); err != nil {
		log.Printf("[%s] warning: could not save token: %v", h.PluginName, err)
	}

	return tok, nil
}

// saveTokenFile persists the token for next run
func (h *OAuthHandler) saveTokenFile(tok *oauth2.Token) error {
	if h.TokenFile == "" {
		return fmt.Errorf("no token file configured")
	}
	if err := os.MkdirAll(filepath.Dir(h.TokenFile), 0700); err != nil {
		return err
	}
	f, err := os.Create(h.TokenFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(tok)
}

// NewGoogleOAuth creates an OAuth handler for Google APIs
func NewGoogleOAuth(credentialsFile, tokenFile, pluginName string, scopes []string) (*OAuthHandler, error) {
	b, err := os.ReadFile(credentialsFile)
	if err != nil {
		return nil, fmt.Errorf("read credentials: %w", err)
	}

	cfg, err := google.ConfigFromJSON(b, scopes...)
	if err != nil {
		return nil, fmt.Errorf("parse credentials: %w", err)
	}

	return &OAuthHandler{
		Config:          cfg,
		TokenFile:       tokenFile,
		CredentialsFile: credentialsFile,
		Scopes:          scopes,
		PluginName:      pluginName,
	}, nil
}
