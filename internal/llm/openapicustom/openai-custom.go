package openapicustom

import (
	"context"
	"encoding/json"
	"errors"

	"bible-api-service/internal/llm/provider"
	"github.com/gofor-little/env"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAICustomClient struct {
	llm llms.Model
}

func NewOpenAICustom(llm llms.Model) provider.LLMClient {
	return &OpenAICustomClient{llm: llm}
}

func NewClient() (provider.LLMClient, error) {
	apiKey, err := env.MustGet("OPENAI_CUSTOM_API_KEY")
	if err != nil {
		return nil, errors.New("OPENAI_CUSTOM_API_KEY environment variable not set")
	}

	baseURL, err := env.MustGet("OPENAI_CUSTOM_BASE_URL")
	if err != nil {
		return nil, errors.New("OPENAI_CUSTOM_BASE_URL environment variable not set")
	}

	llm, err := openai.New(openai.WithToken(apiKey), openai.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	return NewOpenAICustom(llm), nil
}

func (c *OpenAICustomClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, error) {
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
		llms.WithToolChoice("any"),
	)
	if err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, nil
}
