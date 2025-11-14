package config

import (
	"log"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/retriever/fileretriever"
)

func InitFeatureFlags() {
	err := gofeatureflag.Init(gofeatureflag.Config{
		PollingInterval: 10, // 10 seconds
		Retriever: &fileretriever.Retriever{
			Path: "configs/flags.yaml",
		},
	})
	if err != nil {
		log.Fatalf("Error while initializing go-feature-flag: %v", err)
	}
}
