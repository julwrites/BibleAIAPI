package config

import (
	"log"
	"os"

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
		PollingInterval: 10, // 10 seconds
		Retriever: &FallbackRetriever{
			Primary: &githubretriever.Retriever{
				RepositorySlug: "julwrites/BibleAIAPI",
				Branch:         "main",
				FilePath:       "configs/flags.yaml",
				GithubToken:    githubToken,
			},
			Secondary: &fileretriever.Retriever{
				Path: "configs/flags.yaml",
			},
		},
		FileFormat: "yaml",
		Logger:     log.New(os.Stdout, "", log.LstdFlags),
	})
	if err != nil {
		log.Fatalf("Error while initializing go-feature-flag: %v", err)
	}
}
