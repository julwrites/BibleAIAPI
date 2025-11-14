package llm

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

type OpenAIClient struct {
	llm *openai.LLM
}

func NewOpenAIClient() (*OpenAIClient, error) {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return nil, errors.New("OPENAI_API_KEY environment variable not set")
	}
	llm, err := openai.New()
	if err != nil {
		return nil, err
	}
	return &OpenAIClient{llm: llm}, nil
}

func (c *OpenAIClient) Query(ctx context.Context, prompt string, schemaJSON string) (string, error) {
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
