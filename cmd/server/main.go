package main

import (
	"bible-api-service/internal/config"
	"bible-api-service/internal/handlers"
	"bible-api-service/internal/middleware"
	"log"
	"net/http"
	"os"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
)

func main() {
	config.InitFeatureFlags()
	defer gofeatureflag.Close()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	queryHandler := http.HandlerFunc(handlers.QueryHandler)
	http.Handle("/query", middleware.Logging(middleware.APIKeyAuth(queryHandler)))

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
