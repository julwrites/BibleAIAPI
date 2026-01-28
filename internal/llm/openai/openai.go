package openai

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

type OpenAIClient struct {
	llm llms.Model
}

func NewOpenAI(llm llms.Model) provider.LLMClient {
	return &OpenAIClient{llm: llm}
}

func NewClient(ctx context.Context, secretsClient secrets.Client) (provider.LLMClient, error) {
	apiKey, err := secrets.Get(ctx, secretsClient, "OPENAI_API_KEY")
	if err != nil {
		return nil, errors.New("OPENAI_API_KEY secret or environment variable not set")
	}

	llm, err := openai.New(openai.WithToken(apiKey))
	if err != nil {
		return nil, err
	}
	return NewOpenAI(llm), nil
}

func (c *OpenAIClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", "openai", err
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
		return "", "openai", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", "openai", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, "openai", nil
}

func (c *OpenAIClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
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
			log.Printf("openai: stream generation failed: %v", err)
		}
	}()

	return ch, "openai", nil
}

func (c *OpenAIClient) Name() string {
	return "openai"
}
