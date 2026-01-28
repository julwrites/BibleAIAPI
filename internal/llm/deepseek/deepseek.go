package deepseek

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

type DeepseekClient struct {
	llm llms.Model
}

func NewDeepseek(llm llms.Model) provider.LLMClient {
	return &DeepseekClient{llm: llm}
}

func NewClient(ctx context.Context, secretsClient secrets.Client) (provider.LLMClient, error) {
	apiKey, err := secrets.Get(ctx, secretsClient, "DEEPSEEK_API_KEY")
	if err != nil {
		return nil, errors.New("DEEPSEEK_API_KEY secret or environment variable not set")
	}

	baseURL := "https://api.deepseek.com"

	llm, err := openai.New(openai.WithToken(apiKey), openai.WithBaseURL(baseURL))
	if err != nil {
		return nil, err
	}
	return NewDeepseek(llm), nil
}

func (c *DeepseekClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, string, error) {
	var toolSchema llms.FunctionDefinition
	if err := json.Unmarshal([]byte(schemaJSON), &toolSchema); err != nil {
		return "", "deepseek", err
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
		return "", "deepseek", err
	}

	if len(completion.Choices) == 0 || len(completion.Choices[0].ToolCalls) == 0 {
		return "", "deepseek", errors.New("no tool call found in LLM response")
	}

	return completion.Choices[0].ToolCalls[0].FunctionCall.Arguments, "deepseek", nil
}

func (c *DeepseekClient) Stream(ctx context.Context, prompt string) (<-chan string, string, error) {
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

		if _, err := c.llm.GenerateContent(ctx,
			messages,
			llms.WithStreamingFunc(func(ctx context.Context, chunk []byte) error {
				ch <- string(chunk)
				return nil
			}),
		); err != nil {
			log.Printf("deepseek: stream generation failed: %v", err)
		}
	}()

	return ch, "deepseek", nil
}

func (c *DeepseekClient) Name() string {
	return "deepseek"
}
