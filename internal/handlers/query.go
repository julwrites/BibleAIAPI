package handlers

import (
	"bible-api-service/internal/biblegateway"
	"bible-api-service/internal/llm"
	"bible-api-service/internal/util"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	gofeatureflag "github.com/thomaspoignant/go-feature-flag"
	"github.com/thomaspoignant/go-feature-flag/ffcontext"
)

type QueryRequest struct {
	Query struct {
		Verses []string `json:"verses"`
		Words  []string `json:"words"`
		OQuery string   `json:"oquery"`
	} `json:"query"`
	Context struct {
		Instruction string   `json:"instruction"`
		PQuery      []string `json:"pquery"`
		Verses      []string `json:"verses"`
		Words       []string `json:"words"`
		User        struct {
			Version string `json:"version"`
		} `json:"user"`
	} `json:"context"`
}

func QueryHandler(w http.ResponseWriter, r *http.Request) {
	var request QueryRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		util.JSONError(w, http.StatusBadRequest, "Invalid request payload")
		return
	}

	if request.Context.User.Version == "" {
		request.Context.User.Version = "ESV"
	}

	if request.Context.Instruction != "" {
		handleInstruction(w, r, request)
	} else {
		handleDirectQuery(w, r, request)
	}
}

func handleDirectQuery(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	if len(request.Query.Verses) > 0 {
		var verseText []string
		for _, verseRef := range request.Query.Verses {
			verse, err := biblegateway.GetVerse(verseRef, request.Context.User.Version)
			if err != nil {
				util.JSONError(w, http.StatusInternalServerError, "Failed to get verse")
				return
			}
			verseText = append(verseText, verse.Text)
		}
		json.NewEncoder(w).Encode(map[string]string{"verse": strings.Join(verseText, "\n")})
	} else if len(request.Query.Words) > 0 {
		var allResults []biblegateway.SearchResult
		for _, word := range request.Query.Words {
			results, err := biblegateway.SearchWords(word, request.Context.User.Version)
			if err != nil {
				util.JSONError(w, http.StatusInternalServerError, "Failed to search words")
				return
			}
			allResults = append(allResults, results...)
		}
		json.NewEncoder(w).Encode(allResults)
	} else if request.Query.OQuery != "" {
		llmClient, err := llm.GetClient()
		if err != nil {
			util.JSONError(w, http.StatusInternalServerError, err.Error())
			return
		}

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
	} else {
		util.JSONError(w, http.StatusBadRequest, "No query provided")
	}
}

func handleInstruction(w http.ResponseWriter, r *http.Request, request QueryRequest) {
	evalCtx := ffcontext.NewEvaluationContext("anonymous")
	instructionData, err := gofeatureflag.JSONVariation(request.Context.Instruction, evalCtx, map[string]interface{}{})
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

	llmClient, err := llm.GetClient()
	if err != nil {
		util.JSONError(w, http.StatusInternalServerError, err.Error())
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
