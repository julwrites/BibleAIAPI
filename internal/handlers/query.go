package handlers

import (
	"bible-api-service/internal/bible"
	"bible-api-service/internal/bible/providers/biblecom"
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
	"fmt"
	"log"
	"net/http"
	"strings"
)

// QueryHandler is the main handler for the /query endpoint.
type QueryHandler struct {
	ProviderManager *bible.ProviderManager
	GetLLMClient    GetLLMClient
	FFClient        FFClient
	ChatService     ChatService
	VersionManager  *bible.VersionManager
}

// NewQueryHandler creates a new QueryHandler with default clients.
func NewQueryHandler(secretsClient secrets.Client, versionManager *bible.VersionManager) *QueryHandler {
	// Initialize providers
	gatewayProvider := biblegateway.NewScraper()
	hubProvider := biblehub.NewScraper()
	nowProvider := biblenow.NewScraper()
	comProvider := biblecom.NewScraper()

	// Initialize ProviderManager with default/primary (gateway)
	bibleManager := bible.NewProviderManager(gatewayProvider)

	// Register all providers
	bibleManager.RegisterProvider(bible.DefaultProviderName, gatewayProvider)
	bibleManager.RegisterProvider("biblehub", hubProvider)
	bibleManager.RegisterProvider("biblenow", nowProvider)
	bibleManager.RegisterProvider("biblecom", comProvider)

	getLLMClient := func() (provider.LLMClient, error) {
		return llm.NewFallbackClient(context.Background(), secretsClient)
	}
	return &QueryHandler{
		ProviderManager: bibleManager,
		GetLLMClient:    getLLMClient,
		FFClient:        &GoFeatureFlagClient{},
		ChatService:     chat.NewChatService(bibleManager, getLLMClient),
		VersionManager:  versionManager,
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

	// Dynamic Provider Selection
	providerName, providerVersion, err := h.VersionManager.SelectProvider(request.Context.User.Version, nil)
	if err != nil {
		log.Printf("Provider selection failed for %s: %v. Falling back to %s.", request.Context.User.Version, err, bible.DefaultProviderName)
		// Fallback
		providerName = bible.DefaultProviderName
		providerVersion = request.Context.User.Version
	}

	// Update the version in request context to the provider-specific code
	request.Context.User.Version = providerVersion

	if hasPrompt {
		h.handlePromptQuery(w, r, request, providerName)
	} else if hasVerses {
		h.handleVerseQuery(w, r, request, providerName)
	} else if hasWords {
		h.handleWordSearchQuery(w, r, request, providerName)
	}
}

func (h *QueryHandler) handlePromptQuery(w http.ResponseWriter, r *http.Request, request QueryRequest, providerName string) {
	// Validation: Stream and Schema are mutually exclusive
	if request.Options.Stream && request.Context.Schema != "" {
		util.JSONError(w, http.StatusBadRequest, "Stream and Schema are mutually exclusive")
		return
	}

	// Determine schema. If not provided in Context, use default "Open Query" schema.
	// Default schema is ONLY injected if NOT streaming.
	schema := request.Context.Schema
	if !request.Options.Stream && schema == "" {
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
		VerseRefs:  request.Context.Verses,
		Words:      request.Context.Words,
		Version:    request.Context.User.Version, // This is now providerVersion
		Provider:   providerName,
		Prompt:     request.Query.Prompt,
		Schema:     schema,
		AIProvider: request.Context.User.AIProvider,
		Stream:     request.Options.Stream,
		History:    request.Context.History,
	}

	result, err := h.ChatService.Process(r.Context(), chatReq)
	if err != nil {
		log.Printf("ChatService.Process failed: %v", err)
		util.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if result.IsStream {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		flusher, ok := w.(http.Flusher)
		if !ok {
			http.Error(w, "Streaming not supported", http.StatusInternalServerError)
			return
		}

		// Send Meta event
		metaBytes, _ := json.Marshal(result.Meta)
		fmt.Fprintf(w, "event: meta\ndata: %s\n\n", metaBytes)
		flusher.Flush()

		// Stream chunks
		for chunk := range result.Stream {
			data := map[string]string{"delta": chunk}
			dataBytes, _ := json.Marshal(data)
			fmt.Fprintf(w, "event: chunk\ndata: %s\n\n", dataBytes)
			flusher.Flush()
		}

		// Send Done event
		fmt.Fprintf(w, "event: done\ndata: [DONE]\n\n")
		flusher.Flush()

	} else {
		w.Header().Set("Content-Type", "application/json")
		response := map[string]interface{}{
			"data": result.Data,
			"meta": result.Meta,
		}
		json.NewEncoder(w).Encode(response)
	}
}

func (h *QueryHandler) handleVerseQuery(w http.ResponseWriter, r *http.Request, request QueryRequest, providerName string) {
	p, err := h.ProviderManager.GetProvider(providerName)
	if err != nil {
		log.Printf("Failed to get provider %s: %v", providerName, err)
		util.JSONError(w, http.StatusInternalServerError, "Provider configuration error")
		return
	}

	var verseText []string
	for _, verseRef := range request.Query.Verses {
		book, chapter, verseNum, err := util.ParseVerseReference(verseRef)
		if err != nil {
			util.JSONError(w, http.StatusBadRequest, err.Error())
			return
		}

		verse, err := p.GetVerse(book, chapter, verseNum, request.Context.User.Version)
		if err != nil {
			log.Printf("Provider %s GetVerse failed for %s %s:%s: %v", providerName, book, chapter, verseNum, err)
			util.JSONError(w, http.StatusInternalServerError, "Failed to get verse")
			return
		}
		verseText = append(verseText, verse)
	}
	json.NewEncoder(w).Encode(map[string]string{"verse": strings.Join(verseText, "\n")})
}

func (h *QueryHandler) handleWordSearchQuery(w http.ResponseWriter, r *http.Request, request QueryRequest, providerName string) {
	log.Printf("Handling word search query for words: %v using provider: %s", request.Query.Words, providerName)

	p, err := h.ProviderManager.GetProvider(providerName)
	if err != nil {
		log.Printf("Failed to get provider %s: %v", providerName, err)
		util.JSONError(w, http.StatusInternalServerError, "Provider configuration error")
		return
	}

	allResults := make([]bible.SearchResult, 0)
	for _, word := range request.Query.Words {
		results, err := p.SearchWords(word, request.Context.User.Version)
		if err != nil {
			log.Printf("Error searching words '%s' with provider %s: %v", word, providerName, err)
			util.JSONError(w, http.StatusInternalServerError, "Failed to search words")
			return
		}
		log.Printf("Found %d results for word '%s'", len(results), word)
		allResults = append(allResults, results...)
	}
	json.NewEncoder(w).Encode(allResults)
}
