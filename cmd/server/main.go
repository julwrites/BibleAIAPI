package main

import (
	"bible-api-service/internal/config"
	"bible-api-service/internal/handlers"
	"bible-api-service/internal/middleware"
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/storage"
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

	ctx := context.Background()
	projectID := os.Getenv("GCP_PROJECT_ID")

	secretsClient, err := secrets.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("could not create secrets client: %v", err)
	}

	// storageClient with fallback to Mock for local development
	var storageClient storage.Client
	storageClient, err = storage.NewFirestoreClient(ctx, projectID)
	if err != nil {
		log.Printf("could not create firestore client, falling back to in-memory mock: %v", err)
		storageClient = storage.NewMockClient()
	} else {
		defer storageClient.Close()
	}

	authMiddleware := middleware.NewAuthMiddleware(secretsClient, storageClient)

	// Admin Handler
	adminHandler := handlers.NewAdminHandler(storageClient, secretsClient)

	// Register Routes
	queryHandler := handlers.NewQueryHandler(secretsClient)

	http.Handle("/query", middleware.Logging(authMiddleware.APIKeyAuth(queryHandler)))

	// Admin Routes (Publicly accessible, internal auth via password)
	http.Handle("/admin", middleware.Logging(http.HandlerFunc(adminHandler.ServeAdminUI)))
	http.Handle("/api/admin/keys", middleware.Logging(http.HandlerFunc(adminHandler.CreateKey)))

	log.Printf("Server starting on port %s\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("could not listen on port %s %v", port, err)
	}
}
