package llm

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/googleai"
)

type GeminiClient struct {
	llm *googleai.GoogleAI
}

func NewGeminiClient() (*GeminiClient, error) {
	llm, err := googleai.New(context.Background())
	if err != nil {
		return nil, err
	}
	return &GeminiClient{llm: llm}, nil
}

func (c *GeminiClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, error) {
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
