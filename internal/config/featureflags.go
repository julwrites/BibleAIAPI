package config

import (
	"os"
	"log"
	"time"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

func InitFeatureFlags() {
	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Println("GITHUB_TOKEN is not set. Feature flags may not be retrieved from GitHub.")
	}
	err := gofeatureflag.Init(gofeatureflag.Config{
		PollingInterval: 10 * time.Second,
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
