package secrets

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/gax-go/v2"
	secretmanagerpb "google.golang.org/genproto/googleapis/cloud/secretmanager/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type mockSecretManagerClient struct {
	accessSecretVersionFunc func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error)
	closeFunc               func() error
}

func (m *mockSecretManagerClient) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return m.accessSecretVersionFunc(ctx, req, opts...)
}

func (m *mockSecretManagerClient) Close() error {
	return m.closeFunc()
}

func NewTestClient(mock *mockSecretManagerClient, projectID string) Client {
	return &googleSecretManagerClient{
		client:    mock,
		projectID: projectID,
	}
}

func TestGetSecret(t *testing.T) {
	tests := []struct {
		name          string
		secretName    string
		projectID     string
		mockResponse  *secretmanagerpb.AccessSecretVersionResponse
		mockError     error
		expectedValue string
		expectedError error
	}{
		{
			name:       "Successful retrieval",
			secretName: "my-secret",
			projectID:  "my-project",
			mockResponse: &secretmanagerpb.AccessSecretVersionResponse{
				Payload: &secretmanagerpb.SecretPayload{
					Data: []byte("my-secret-value"),
				},
			},
			mockError:     nil,
			expectedValue: "my-secret-value",
			expectedError: nil,
		},
		{
			name:          "Secret not found",
			secretName:    "my-secret",
			projectID:     "my-project",
			mockResponse:  nil,
			mockError:     status.Error(codes.NotFound, "secret not found"),
			expectedValue: "",
			expectedError: fmt.Errorf("failed to access secret version: %v", status.Error(codes.NotFound, "secret not found")),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockSecretManagerClient{
				accessSecretVersionFunc: func(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest, opts ...gax.CallOption) (*secretmanagerpb.AccessSecretVersionResponse, error) {
					return tt.mockResponse, tt.mockError
				},
				closeFunc: func() error { return nil },
			}
			client := NewTestClient(mock, tt.projectID)

			value, err := client.GetSecret(context.Background(), tt.secretName)

			if value != tt.expectedValue {
				t.Errorf("unexpected value: got %q, want %q", value, tt.expectedValue)
			}

			if !cmp.Equal(err, tt.expectedError, cmp.Comparer(func(x, y error) bool {
				if x == nil || y == nil {
					return x == y
				}
				return x.Error() == y.Error()
			})) {
				t.Errorf("unexpected error: got %v, want %v", err, tt.expectedError)
			}
		})
	}
}
