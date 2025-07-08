package api

import (
	"context"
	"fmt"
	"time"

	"github.com/sashabaranov/go-openai"
)

// AskOpenAI sends a prompt to the OpenAI API and retrieves the response.
func AskOpenAI(client *openai.Client, model, prompt string, maxTokens int) (*openai.ChatCompletionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   maxTokens,
			Temperature: 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	return &resp, nil
}

// AskOpenAI sends a prompt to the OpenAI API and retrieves the response.
func AskOpenAIwithRandomInput(client *openai.Client, model string, numWords int, maxTokens int) (*openai.ChatCompletionResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	prompt := generateRandomPhrase(numWords)
	resp, err := client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are a helpful assistant.",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			MaxTokens:   maxTokens,
			Temperature: 1,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API request failed: %w", err)
	}
	return &resp, nil
}

// GetFirstAvailableModel retrieves the first available model from the OpenAI API.
func GetFirstAvailableModel(client *openai.Client) (string, error) {
	modelList, err := client.ListModels(context.Background())
	if err != nil {
		return "", fmt.Errorf("failed to list models: %w", err)
	}

	if len(modelList.Models) == 0 {
		return "", fmt.Errorf("no models available")
	}

	return modelList.Models[0].ID, nil
}

// EstimateInputTokens
func EstimateInputTokens(client *openai.Client, modelName, prompt string, numWords int) (int, error) {
	if numWords > 0 {
		resp, err := AskOpenAIwithRandomInput(client, modelName, numWords/4, 1)
		if err != nil {
			return 0, err
		}
		return resp.Usage.PromptTokens, nil
	}

	resp, err := AskOpenAI(client, modelName, prompt, 1)
	if err != nil {
		return 0, err
	}
	return resp.Usage.PromptTokens, nil
}
