package main

import (
	"bible-api-service/internal/config"
	"bible-api-service/internal/handlers"
	"bible-api-service/internal/middleware"
	"bible-api-service/internal/secrets"
	"context"
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

	secretsClient, err := secrets.NewClient(context.Background(), os.Getenv("GCP_PROJECT_ID"))
	if err != nil {
		log.Fatalf("could not create secrets client: %v", err)
	}
	authMiddleware := middleware.NewAuthMiddleware(secretsClient)

	queryHandler := handlers.NewQueryHandler()
	http.Handle("/query", middleware.Logging(authMiddleware.APIKeyAuth(queryHandler)))

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
