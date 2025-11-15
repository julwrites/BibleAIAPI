package config

import (
	"context"
	"fmt"
	"log"

	"github.com/thomaspoignant/go-feature-flag/retriever"
)

// FallbackRetriever is a custom retriever that falls back to a file if the GitHub retriever fails.
type FallbackRetriever struct {
	Primary   retriever.Retriever
	Secondary retriever.Retriever
}

// NewFallbackRetriever creates a new instance of the FallbackRetriever.
func NewFallbackRetriever(primary, secondary retriever.Retriever) *FallbackRetriever {
	return &FallbackRetriever{
		Primary:   primary,
		Secondary: secondary,
	}
}

// Retrieve attempts to get the flags from the primary retriever and falls back to the secondary.
func (r *FallbackRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	// Try to get the flags from the primary retriever first.
	flags, err := r.Primary.Retrieve(ctx)
	if err == nil {
		log.Println("Successfully retrieved feature flags from primary retriever.")
		return flags, nil
	}

	// If the primary retriever fails, log a warning and try the secondary.
	log.Printf("Failed to retrieve feature flags from primary retriever: %v. Falling back to secondary.", err)
	flags, err = r.Secondary.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve flags from both primary and secondary retrievers: %w", err)
	}

	log.Println("Successfully retrieved feature flags from secondary retriever.")
	return flags, nil
}

// Name returns the name of the retriever.
func (r *FallbackRetriever) Name() string {
	return "fallbackRetriever"
}
