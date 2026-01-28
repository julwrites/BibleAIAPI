package openrouter

import (
	"context"
	"encoding/json"
	"errors"
	"log"

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

func (c *OpenRouterClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", "openrouter", err
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
		return "", "openrouter", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", "openrouter", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, "openrouter", nil
}

func (c *OpenRouterClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
	ch := make(chan string)

	go func() {
		defer close(ch)

		messages := []llms.MessageContent{
			{
				Role: llms.ChatMessageTypeHuman,
				Parts: []llms.ContentPart{
					llms.TextPart(prompt),
				},
			},
		}

		// We ignore the response from GenerateContent because the chunks are sent via the streaming callback.
		// We can't return the error from the goroutine to the caller (who has already received the channel),
		// so we log it.
		if _, err := c.llm.GenerateContent(ctx,
			messages,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				ch <- string(chunk)
				return nil
			}),
		); err != nil {
			log.Printf("openrouter: stream generation failed: %v", err)
		}
	}()

	return ch, "openrouter", nil
}

func (c *OpenRouterClient) Name() string {
	return "openrouter"
}
