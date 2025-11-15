package config

import (
	"log"
	"time"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/githubretriever"
)

func InitFeatureFlags() {
	err := gofeatureflag.Init(gofeatureflag.Config{
		PollingInterval: 10 * time.Second,
		Retriever: &githubretriever.Retriever{
			RepositorySlug: "julwrites/BibleAIAPI",
			FilePath:       "configs/flags.yaml",
		},
	})
	if err != nil {
		log.Fatalf("Error while initializing go-feature-flag: %v", err)
	}
}
