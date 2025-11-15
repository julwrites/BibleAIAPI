package config

import (
	"context"
	"fmt"
	"log"

	"github.com/thomaspoignant/go-feature-flag/retriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

// fallbackRetriever is a custom retriever that falls back to a file if the GitHub retriever fails.
type fallbackRetriever struct {
	githubRetriever retriever.Retriever
	fileRetriever   retriever.Retriever
}

// NewFallbackRetriever creates a new instance of the fallbackRetriever.
func NewFallbackRetriever(githubConfig *githubretriever.Retriever, fileConfig *fileretriever.Retriever) retriever.Retriever {
	return &fallbackRetriever{
		githubRetriever: githubConfig,
		fileRetriever:   fileConfig,
	}
}

// Retrieve attempts to get the flags from GitHub and falls back to the local file.
func (r *fallbackRetriever) Retrieve(ctx context.Context) ([]byte, error) {
	// Try to get the flags from GitHub first.
	flags, err := r.githubRetriever.Retrieve(ctx)
	if err == nil {
		log.Println("Successfully retrieved feature flags from GitHub.")
		return flags, nil
	}

	// If GitHub fails, log a warning and try the file retriever.
	log.Printf("Failed to retrieve feature flags from GitHub: %v. Falling back to local file.", err)
	flags, err = r.fileRetriever.Retrieve(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve flags from both GitHub and local file: %w", err)
	}

	log.Println("Successfully retrieved feature flags from local file.")
	return flags, nil
}

// Name returns the name of the retriever.
func (r *fallbackRetriever) Name() string {
	return "fallbackRetriever"
}
