package openapicustom

import (
	"context"
	"encoding/json"
	"errors"

	"bible-api-service/internal/llm/provider"
	"bible-api-service/internal/secrets"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAICustomClient struct {
	llm llms.Model
}

func NewOpenAICustom(llm llms.Model) provider.LLMClient {
	return &OpenAICustomClient{llm: llm}
}

func NewClient(ctx context.Context, secretsClient secrets.Client) (provider.LLMClient, error) {
	apiKey, err := secrets.Get(ctx, secretsClient, "OPENAI_CUSTOM_API_KEY")
	if err != nil {
		return nil, errors.New("OPENAI_CUSTOM_API_KEY secret or environment variable not set")
	}

	baseURL, err := secrets.Get(ctx, secretsClient, "OPENAI_CUSTOM_BASE_URL")
	if err != nil {
		return nil, errors.New("OPENAI_CUSTOM_BASE_URL secret or environment variable not set")
	}

	llm, err := openai.New(openai.WithToken(apiKey), openai.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	return NewOpenAICustom(llm), nil
}

func (c *OpenAICustomClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", "openai-custom", err
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
		return "", "openai-custom", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", "openai-custom", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, "openai-custom", nil
}

func (c *OpenAICustomClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	return nil, "openai-custom", errors.New("not implemented")
}

func (c *OpenAICustomClient) Name() string {
	return "openai-custom"
}
