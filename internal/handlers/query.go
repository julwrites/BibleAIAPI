package handlers

import (
	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblegateway"
	"bible-api-service/internal/bible/providers/biblehub"
	"bible-api-service/internal/bible/providers/biblenow"
	"bible-api-service/internal/chat"
	"bible-api-service/internal/llm"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"
	"bible-api-service/internal/util"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

// QueryHandler is the main handler for the /query endpoint.
type QueryHandler struct {
	BibleGatewayClient BibleGatewayClient
	GetLLMClient       GetLLMClient
	FFClient           FFClient
	ChatService        ChatService
	VersionManager     *bible.VersionManager
	ProviderName       string
}

// NewQueryHandler creates a new QueryHandler with default clients.
func NewQueryHandler(secretsClient secrets.Client, versionManager *bible.VersionManager) *QueryHandler {
	// Initialize the Bible provider manager based on environment variable
	providerName := os.Getenv("BIBLE_PROVIDER")
	if providerName == "" {
		providerName = "biblegateway"
	}

	var bibleProvider bible.Provider
	switch providerName {
	case "biblehub":
		bibleProvider = biblehub.NewScraper()
	case "biblenow":
		bibleProvider = biblenow.NewScraper()
	default:
		bibleProvider = biblegateway.NewScraper()
	}
	bibleManager := bible.NewProviderManager(bibleProvider)

	getLLMClient := func() (provider.LLMClient, error) {
		return llm.NewFallbackClient(context.Background(), secretsClient)
	}
	return &QueryHandler{
		BibleGatewayClient: bibleManager,
		GetLLMClient:       getLLMClient,
		FFClient:           &GoFeatureFlagClient{},
		ChatService:        chat.NewChatService(bibleManager, getLLMClient),
		VersionManager:     versionManager,
		ProviderName:       providerName,
	}
}

// ServeHTTP handles the HTTP request.
func (h *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	// Validate exactly one of verses, words, or prompt is present
	hasVerses := len(request.Query.Verses) > 0
	hasWords := len(request.Query.Words) > 0
	hasPrompt := request.Query.Prompt != ""

	// Count true values
	count := 0
	if hasVerses {
		count++
	}
	if hasWords {
		count++
	}
	if hasPrompt {
		count++
	}

	if count != 1 {
		util.JSONError(w, http.StatusBadRequest, "Query must contain exactly one of: verses, words, or prompt")
		return
	}

	if !hasPrompt {
		// Check if Context (excluding User) is non-empty
		hasContext := len(request.Context.History) > 0 ||
			request.Context.Schema != "" ||
			len(request.Context.Verses) > 0 ||
			len(request.Context.Words) > 0

		if hasContext {
			util.JSONError(w, http.StatusBadRequest, "Context object (excluding user preferences) is only valid with a prompt query")
			return
		}
	}

	if request.Context.User.Version == "" {
		request.Context.User.Version = "ESV"
	}

	// Resolve provider-specific version
	providerVersion, err := h.VersionManager.GetProviderCode(request.Context.User.Version, h.ProviderName)
	if err != nil {
		log.Printf("Version lookup failed for %s: %v", request.Context.User.Version, err)
		// Fallback to unified version if lookup fails
		providerVersion = request.Context.User.Version
	}
	// Temporarily override the version in request context so handlers use the provider code
	// Note: Use a separate variable or modify request object if safe.
	// We'll pass providerVersion explicitly where needed or modify the request struct locally.
	request.Context.User.Version = providerVersion

	if hasPrompt {
		h.handlePromptQuery(w, r, request)
	} else if hasVerses {
		h.handleVerseQuery(w, r, request)
	} else if hasWords {
		h.handleWordSearchQuery(w, r, request)
	}
}

func (h *QueryHandler) handlePromptQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	// Determine schema. If not provided in Context, use default "Open Query" schema.
	schema := request.Context.Schema
	if schema == "" {
		schema = `{
			"name": "oquery_response",
			"description": "A response to an open-ended query.",
			"parameters": {
				"type": "object",
				"properties": {
					"text": {
						"type": "string",
						"description": "The response to the query in semantic HTML format."
					},
					"references": {
						"type": "array",
						"items": {
							"type": "object",
							"properties": {
								"verse": {
									"type": "string",
									"description": "A relevant Bible verse reference."
								},
								"url": {
									"type": "string",
									"description": "A URL to the verse on Bible Gateway."
								}
							}
						}
					}
				}
			}
		}`
	}

	chatReq := chat.Request{
		VerseRefs: request.Context.Verses,
		Words:     request.Context.Words,
		Version:   request.Context.User.Version, // This is now providerVersion
		Prompt:    request.Query.Prompt,
		Schema:    schema,
	}

	result, err := h.ChatService.Process(r.Context(), chatReq)
	if err != nil {
		log.Printf("ChatService.Process failed: %v", err)
		util.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *QueryHandler) handleVerseQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	var verseText []string
	for _, verseRef := range request.Query.Verses {
		book, chapter, verseNum, err := util.ParseVerseReference(verseRef)
		if err != nil {
			util.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		verse, err := h.BibleGatewayClient.GetVerse(book, chapter, verseNum, request.Context.User.Version)
		if err != nil {
			log.Printf("BibleGatewayClient.GetVerse failed for %s %s:%s: %v", book, chapter, verseNum, err)
			util.JSONError(w, http.StatusInternalServerError, "Failed to get verse")
			return
		}
		verseText = append(verseText, verse)
	}
	json.NewEncoder(w).Encode(map[string]string{"verse": strings.Join(verseText, "\n")})
}

func (h *QueryHandler) handleWordSearchQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	log.Printf("Handling word search query for words: %v", request.Query.Words)
	allResults := make([]bible.SearchResult, 0)
	for _, word := range request.Query.Words {
		results, err := h.BibleGatewayClient.SearchWords(word, request.Context.User.Version)
		if err != nil {
			log.Printf("Error searching words '%s': %v", word, err)
			util.JSONError(w, http.StatusInternalServerError, "Failed to search words")
			return
		}
		log.Printf("Found %d results for word '%s'", len(results), word)
		allResults = append(allResults, results...)
	}
	json.NewEncoder(w).Encode(allResults)
}
