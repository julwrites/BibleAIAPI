package secrets

import (
	"context"
	"errors"
	"testing"

	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
)

type mockSecretManagerClient struct {
	accessSecretVersionFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	closeFunc               func() error
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	if m.accessSecretVersionFunc != nil {
		return m.accessSecretVersionFunc(ctx, req, opts...)
	}
	return nil, nil
}

func (m *mockSecretManagerClient) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func TestNew(t *testing.T) {
	mock := &mockSecretManagerClient{
		closeFunc: func() error { return nil },
	}
	client := New(mock, "test-project")
	if client == nil {
		t.Error("expected client to not be nil")
	}
}

func TestNewClient(t *testing.T) {
	// This test verifies that the NewClient function falls back to the
	// environment variable client when it fails to create a Google Secret
	// environment where credentials are not available.
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "/nonexistent")
	// If GCP_PROJECT_ID is not set, it defaults to env client.
	// If it IS set, it tries to create secret manager client.
	// We want to test the fallback when secret manager creation fails (if possible to mock that without failing NewClient itself?)
	// Actually secretmanager.NewClient(ctx) usually succeeds unless auth fails significantly or something.
	// But without creds, it might not fail immediately?
	// The logs say "failed to create secret manager client" in the CI run when open /nonexistent fails.

	client, err := NewClient(context.Background(), "test-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// It should return an *EnvClient struct pointer but implementing Client interface
	if _, ok := client.(*EnvClient); !ok {
		t.Errorf("expected an env client, but got %T", client)
	}

	// Test case where Project ID is empty
	clientEmpty, err := NewClient(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := clientEmpty.(*EnvClient); !ok {
		t.Errorf("expected an env client for empty project ID, but got %T", clientEmpty)
	}
}

func TestGoogleSecretManagerClient_GetSecret(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		mock := &mockSecretManagerClient{
			accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
				if req.Name != "projects/test-project/secrets/my-secret/versions/latest" {
					return nil, errors.New("invalid secret name")
				}
				return &secretmanagerpb.AccessSecretVersionResponse{
					Payload: &secretmanagerpb.SecretPayload{
						Data: []byte("secret-value"),
					},
				}, nil
			},
		}

		client := New(mock, "test-project")
		val, err := client.GetSecret(context.Background(), "my-secret")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if val != "secret-value" {
			t.Errorf("expected 'secret-value', got '%s'", val)
		}
	})

	t.Run("Failure", func(t *testing.T) {
		mock := &mockSecretManagerClient{
			accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
				return nil, errors.New("gcp error")
			},
		}

		client := New(mock, "test-project")
		_, err := client.GetSecret(context.Background(), "my-secret")
		if err == nil {
			t.Error("expected error")
		}
	})
}
