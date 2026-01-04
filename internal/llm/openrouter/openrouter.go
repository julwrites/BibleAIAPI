package openrouter

import (
	"context"
	"encoding/json"
	"errors"

	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenRouterClient struct {
	llm llms.Model
}

func NewOpenRouter(llm llms.Model) provider.LLMClient {
	return &OpenRouterClient{llm: llm}
}

func NewClient(ctx context.Context, secretsClient secrets.Client, model string) (provider.LLMClient, error) {
	apiKey, err := secrets.Get(ctx, secretsClient, "OPENROUTER_API_KEY")
	if err != nil {
		return nil, errors.New("OPENROUTER_API_KEY secret or environment variable not set")
	}

	baseURL := "https://openrouter.ai/api/v1"

	var opts []openai.Option
	opts = append(opts, openai.WithToken(apiKey), openai.WithBaseURL(baseURL))
	if model != "" {
		opts = append(opts, openai.WithModel(model))
	}

	llm, err := openai.New(opts...)
	if err != nil {
		return nil, err
	}
	return NewOpenRouter(llm), nil
}

func (c *OpenRouterClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", err
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
		return "", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, nil
}