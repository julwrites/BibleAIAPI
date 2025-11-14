package secrets

import (
	"context"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

// googleSecretManagerClient is a client for interacting with Google Secret Manager.
type googleSecretManagerClient struct {
	client    *secretmanager.Client
	projectID string
}

// NewClient creates a new Secret Manager client.
func NewClient(ctx context.Context, projectID string) (Client, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create secret manager client: %v", err)
	}
	return &googleSecretManagerClient{client: client, projectID: projectID}, nil
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
