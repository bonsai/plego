package core

import "context"

// Authorizer is implemented by plugins that require authentication.
type Authorizer interface {
	// InitAuth initializes authentication before the plugin is used.
	// Should return nil if auth is successful, or handle fallback internally.
	InitAuth(ctx context.Context) error
}
