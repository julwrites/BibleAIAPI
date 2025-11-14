package secrets

import "context"

// Client is an interface for a secrets client.
type Client interface {
	GetSecret(ctx context.Context, name string) (string, error)
}
