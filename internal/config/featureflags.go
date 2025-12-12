package config

import (
	"log"
	"os"
	"time"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

// getPollingInterval returns the polling interval from the environment variable
// FEATURE_FLAG_POLLING_INTERVAL or defaults to 5 minutes.
func getPollingInterval() time.Duration {
	envVal := os.Getenv("FEATURE_FLAG_POLLING_INTERVAL")
	if envVal == "" {
		return 300 * time.Second
	}

	duration, err := time.ParseDuration(envVal)
	if err != nil {
		log.Printf("Invalid FEATURE_FLAG_POLLING_INTERVAL '%s', defaulting to 5 minutes: %v", envVal, err)
		return 300 * time.Second
	}
	return duration
}

func InitFeatureFlags() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Println("GITHUB_TOKEN is not set. Feature flags may not be retrieved from GitHub.")
	}

	pollingInterval := getPollingInterval()
	log.Printf("Feature Flag Polling Interval set to: %v", pollingInterval)

	err := gofeatureflag.Init(gofeatureflag.Config{
		PollingInterval: pollingInterval,
		Retriever: NewFallbackRetriever(
			&githubretriever.Retriever{
				RepositorySlug: "julwrites/BibleAIAPI",
				Branch:         "main",
				FilePath:       "configs/flags.yaml",
				GithubToken:    githubToken,
			},
			&fileretriever.Retriever{
				Path: "configs/flags.yaml",
			},
		),
	})
	if err != nil {
		log.Fatalf("Error while initializing go-feature-flag: %v", err)
	}
}
