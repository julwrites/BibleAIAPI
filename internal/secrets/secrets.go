package secrets

import (
	"context"
	"fmt"
	"log"
	"os"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// secretManagerClient is an interface that wraps the AccessSecretVersion method.
type secretManagerClient interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

// googleSecretManagerClient is a client for interacting with Google Secret Manager.
type googleSecretManagerClient struct {
	client    secretManagerClient
	projectID string
}

func New(client secretManagerClient, projectID string) Client {
	return &googleSecretManagerClient{client: client, projectID: projectID}
}

// NewClient creates a new Secret Manager client. It first attempts to connect
// to Google Secret Manager. If that fails, it falls back to reading secrets
// from environment variables.
func NewClient(ctx context.Context, projectID string) (Client, error) {
	if projectID == "" {
		log.Println("GCP_PROJECT_ID is not set; falling back to environment variables")
		return &EnvClient{}, nil
	}

	c, err := secretmanager.NewClient(ctx)
	if err != nil {
		log.Printf("failed to create secret manager client, falling back to environment variables: %v", err)
		return &EnvClient{}, nil
	}

	return New(c, projectID), nil
}

// GetSecret retrieves a secret from Secret Manager.
func (c *googleSecretManagerClient) GetSecret(ctx context.Context, name string) (string, error) {
	fullName := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", c.projectID, name)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: fullName,
	}

	result, err := c.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to access secret version: %v", err)
	}

	return string(result.Payload.Data), nil
}

// Get retrieves a secret using the provided client, falling back to environment variables if retrieval fails.
func Get(ctx context.Context, client Client, key string) (string, error) {
	val, err := client.GetSecret(ctx, key)
	if err == nil && val != "" {
		return val, nil
	}

	if valEnv := os.Getenv(key); valEnv != "" {
		return valEnv, nil
	}

	if err != nil {
		return "", err
	}
	return "", fmt.Errorf("secret %q not found", key)
}
