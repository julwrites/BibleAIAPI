package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm"
	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/util"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

// QueryHandler is the main handler for the /query endpoint.
type QueryHandler struct {
	BibleGatewayClient BibleGatewayClient
	GetLLMClient       GetLLMClient
	FFClient           FFClient
}

// NewQueryHandler creates a new QueryHandler with default clients.
func NewQueryHandler() *QueryHandler {
	return &QueryHandler{
		BibleGatewayClient: biblegateway.NewScraper(),
		GetLLMClient: func() (provider.LLMClient, error) {
			return llm.NewFallbackClient()
		},
		FFClient: &GoFeatureFlagClient{},
	}
}

// ServeHTTP handles the HTTP request.
func (h *QueryHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var request QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.Context.User.Version == "" {
		request.Context.User.Version = "ESV"
	}

	if request.Context.Instruction != "" {
		h.handleInstruction(w, r, request)
	} else {
		h.handleDirectQuery(w, r, request)
	}
}

func (h *QueryHandler) handleDirectQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	if len(request.Query.Verses) > 0 {
		h.handleVerseQuery(w, r, request)
	} else if len(request.Query.Words) > 0 {
		h.handleWordSearchQuery(w, r, request)
	} else if request.Query.OQuery != "" {
		h.handleOpenQuery(w, r, request)
	} else {
		util.JSONError(w, http.StatusBadRequest, "No query provided")
	}
}

func (h *QueryHandler) handleVerseQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	var verseText []string
	for _, verseRef := range request.Query.Verses {
		parts := strings.Split(verseRef, " ")
		book := parts[0]
		chapterAndVerse := strings.Split(parts[1], ":")
		chapter := chapterAndVerse[0]
		verseNum := chapterAndVerse[1]

		verse, err := h.BibleGatewayClient.GetVerse(book, chapter, verseNum, request.Context.User.Version)
		if err != nil {
			util.JSONError(w, http.StatusInternalServerError, "Failed to get verse")
			return
		}
		verseText = append(verseText, verse)
	}
	json.NewEncoder(w).Encode(map[string]string{"verse": strings.Join(verseText, "\n")})
}

func (h *QueryHandler) handleWordSearchQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	var allResults []biblegateway.SearchResult
	for _, word := range request.Query.Words {
		results, err := h.BibleGatewayClient.SearchWords(word, request.Context.User.Version)
		if err != nil {
			util.JSONError(w, http.StatusInternalServerError, "Failed to search words")
			return
		}
		allResults = append(allResults, results...)
	}
	json.NewEncoder(w).Encode(allResults)
}

func (h *QueryHandler) handleOpenQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	schema := `{
		"name": "oquery_response",
		"description": "A response to an open-ended query.",
		"parameters": {
			"type": "object",
			"properties": {
				"text": {
					"type": "string",
					"description": "The text response to the query."
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
	llmClient, err := h.GetLLMClient()
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	response, err := llmClient.Query(r.Context(), request.Query.OQuery, schema)
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to query LLM")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to parse LLM response")
		return
	}
	json.NewEncoder(w).Encode(result)
}

func (h *QueryHandler) handleInstruction(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	evalCtx := ffcontext.NewEvaluationContext("anonymous")
	instructionData, err := h.FFClient.JSONVariation(request.Context.Instruction, evalCtx, map[string]interface{}{})
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to evaluate feature flag")
		return
	}

	promptTemplate, ok := instructionData["prompt"].(string)
	if !ok {
		util.JSONError(w, http.StatusInternalServerError, "Invalid prompt in feature flag")
		return
	}
	schema, ok := instructionData["schema"].(string)
	if !ok {
		util.JSONError(w, http.StatusInternalServerError, "Invalid schema in feature flag")
		return
	}

	// Collate context
	pquery := append(request.Context.PQuery, request.Query.OQuery)
	verses := append(request.Context.Verses, request.Query.Verses...)
	words := append(request.Context.Words, request.Query.Words...)

	// Prepare data for the template
	templateData := struct {
		Verses []string
		Words  []string
		PQuery []string
	}{
		Verses: verses,
		Words:  words,
		PQuery: pquery,
	}

	// Create and execute the template
	tmpl, err := template.New("prompt").Parse(promptTemplate)
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to parse prompt template")
		return
	}
	var processedPrompt bytes.Buffer
	if err := tmpl.Execute(&processedPrompt, templateData); err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to execute prompt template")
		return
	}

	llmClient, err := h.GetLLMClient()
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, err.Error())
		return
	}
	response, err := llmClient.Query(r.Context(), processedPrompt.String(), schema)
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to query LLM")
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(response), &result); err != nil {
		util.JSONError(w, http.StatusInternalServerError, "Failed to parse LLM response")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
