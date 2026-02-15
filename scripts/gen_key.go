package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
)

func main() {
	clientID := flag.String("client", "", "Client ID for the new key (optional)")
	flag.Parse()

	if *clientID == "" {
		// Generate a random client ID suffix
		b := make([]byte, 4)
		if _, err := rand.Read(b); err != nil {
			fmt.Fprintf(os.Stderr, "Error generating random client ID: %v\n", err)
			os.Exit(1)
		}
		*clientID = fmt.Sprintf("test-client-%s", hex.EncodeToString(b))
	}

	// Generate 32-byte API Key
	keyBytes := make([]byte, 32)
	if _, err := rand.Read(keyBytes); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating API key: %v\n", err)
		os.Exit(1)
	}
	apiKey := hex.EncodeToString(keyBytes)

	fmt.Printf("Generated API Key for client '%s':\n", *clientID)
	fmt.Println(apiKey)
	fmt.Println("\nTo use this key locally, add the following to your environment variables or .env file:")

	// Create JSON snippet
	existingKeys := os.Getenv("API_KEYS")
	var keyMap map[string]string
	if existingKeys != "" {
		if err := json.Unmarshal([]byte(existingKeys), &keyMap); err != nil {
			// If existing keys are invalid, start fresh
			keyMap = make(map[string]string)
		}
	} else {
		keyMap = make(map[string]string)
	}
	keyMap[*clientID] = apiKey

	jsonBytes, err := json.Marshal(keyMap)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error marshaling JSON: %v\n", err)
		os.Exit(1)
	}

	// Escape quotes for shell
	jsonStr := string(jsonBytes)
	escapedJSON := strings.ReplaceAll(jsonStr, "\"", "\\\"")

	fmt.Printf("\nexport API_KEYS=\"%s\"\n", escapedJSON)

	fmt.Println("\nOr in your .env file (if using godotenv or Docker):")
	fmt.Printf("API_KEYS='%s'\n", jsonStr)
}
