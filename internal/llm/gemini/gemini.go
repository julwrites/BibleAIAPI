package gemini

import (
	"context"
	"encoding/json"
	"errors"

	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type GeminiClient struct {
	llm llms.Model
}

func NewGemini(llm llms.Model) provider.LLMClient {
	return &GeminiClient{llm: llm}
}

func NewClient(ctx context.Context, secretsClient secrets.Client) (provider.LLMClient, error) {
	apiKey, err := secrets.Get(ctx, secretsClient, "GEMINI_API_KEY")
	if err != nil {
		return nil, errors.New("GEMINI_API_KEY secret or environment variable not set")
	}

	llm, err := googleai.New(ctx, googleai.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	return NewGemini(llm), nil
}

func (c *GeminiClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", "gemini", err
	}

	messages := []llms.MessageContent{
		{
			Role: llms.ChatMessageTypeHuman,
			Parts: []llms.ContentPart{
				llms.TextPart(prompt),
			},
		},
	}

	completion, err := c.llm.GenerateContent(ctx,
		messages,
		llms.WithTools([]llms.Tool{
			{
				Type:     "function",
				Function: &toolSchema,
			},
		}),
		llms.WithToolChoice("required"),
	)
	if err != nil {
		return "", "gemini", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", "gemini", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, "gemini", nil
}

func (c *GeminiClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	return nil, "gemini", errors.New("not implemented")
}

func (c *GeminiClient) Name() string {
	return "gemini"
}
