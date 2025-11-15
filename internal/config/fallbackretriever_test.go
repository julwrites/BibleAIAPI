package config

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/thomaspoignant/go-feature-flag/retriever"
)

type mockRetriever struct {
	retrieveFunc func(context.Context) ([]byte, error)
}

func (m *mockRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	return m.retrieveFunc(ctx)
}

func (m *mockRetriever) Shutdown(ctx context.Context) error {
	return nil
}

func (m *mockRetriever) Status() retriever.Status {
	return retriever.RetrieverReady
}

func TestFallbackRetriever_Retrieve(t *testing.T) {
	tests := []struct {
		name              string
		primaryResponse   []byte
		primaryError      error
		secondaryResponse []byte
		secondaryError    error
		expectedResponse  []byte
		expectedError     error
	}{
		{
			name:              "Primary succeeds",
			primaryResponse:   []byte("primary"),
			primaryError:      nil,
			secondaryResponse: []byte("secondary"),
			secondaryError:    nil,
			expectedResponse:  []byte("primary"),
			expectedError:     nil,
		},
		{
			name:              "Primary fails, secondary succeeds",
			primaryResponse:   nil,
			primaryError:      errors.New("primary failed"),
			secondaryResponse: []byte("secondary"),
			secondaryError:    nil,
			expectedResponse:  []byte("secondary"),
			expectedError:     nil,
		},
		{
			name:              "Both fail",
			primaryResponse:   nil,
			primaryError:      errors.New("primary failed"),
			secondaryResponse: nil,
			secondaryError:    errors.New("secondary failed"),
			expectedResponse:  nil,
			expectedError:     errors.New("secondary failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			primary := &mockRetriever{
				retrieveFunc: func(context.Context) ([]byte, error) {
					return tt.primaryResponse, tt.primaryError
				},
			}
			secondary := &mockRetriever{
				retrieveFunc: func(context.Context) ([]byte, error) {
					return tt.secondaryResponse, tt.secondaryError
				},
			}
			retriever := &FallbackRetriever{
				Primary:   primary,
				Secondary: secondary,
			}

			response, err := retriever.Retrieve(context.Background())

			if !cmp.Equal(response, tt.expectedResponse) {
				t.Errorf("unexpected response: got %q, want %q", response, tt.expectedResponse)
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
