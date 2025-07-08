package api

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/sashabaranov/go-openai"
)

// MeasureTTFT calculates the maximum and minimum Time to First Token (TTFT) for API responses.
func MeasureTTFT(client *openai.Client, model, prompt string, concurrency int) (float64, float64) {
	var wg sync.WaitGroup
	ttftChan := make(chan float64, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()

			start := time.Now()
			stream, err := client.CreateChatCompletionStream(
				ctx, // ← 使用带超时的context
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
					MaxTokens:   512,
					Temperature: 1,
					Stream:      true,
				},
			)
			if err != nil {
				log.Printf("TTFT Request error (goroutine %d): %v", index, err)
				ttftChan <- -1 // ← 发送错误标记
				return
			}
			defer stream.Close()

			// Listen for the first response
			_, err = stream.Recv()
			if err != nil {
				log.Printf("TTFT Stream error (goroutine %d): %v", index, err)
				ttftChan <- -1 // ← 发送错误标记
				return
			}

			// Record TTFT
			ttft := time.Since(start).Seconds()
			ttftChan <- ttft
		}(i)
	}

	wg.Wait()
	close(ttftChan)

	// Calculate maximum and minimum TTFT
	maxTTFT := 0.0
	minTTFT := math.Inf(1)
	validCount := 0
	for ttft := range ttftChan {
		if ttft < 0 {
			// Skip error markers
			continue
		}
		validCount++
		if ttft > maxTTFT {
			maxTTFT = ttft
		}
		if ttft < minTTFT {
			minTTFT = ttft
		}
	}

	// If no valid responses, return default values
	if validCount == 0 {
		log.Printf("Warning: No valid TTFT measurements for concurrency %d", concurrency)
		return 0, 0
	}

	return maxTTFT, minTTFT
}

func MeasureTTFTwithRandomInput(client *openai.Client, model string, numWords int, concurrency int) (float64, float64) {
	var wg sync.WaitGroup
	ttftChan := make(chan float64, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
			defer cancel()

			prompt := generateRandomPhrase(numWords)
			start := time.Now()
			stream, err := client.CreateChatCompletionStream(
				ctx, // ← 使用带超时的context
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
					MaxTokens:   512,
					Temperature: 1,
					Stream:      true,
				},
			)
			if err != nil {
				log.Printf("TTFT Request error (goroutine %d): %v", index, err)
				ttftChan <- -1 // ← 发送错误标记
				return
			}
			defer stream.Close()

			// Listen for the first response
			_, err = stream.Recv()
			if err != nil {
				log.Printf("TTFT Stream error (goroutine %d): %v", index, err)
				ttftChan <- -1 // ← 发送错误标记
				return
			}

			// Record TTFT
			ttft := time.Since(start).Seconds()
			ttftChan <- ttft
		}(i)
	}

	wg.Wait()
	close(ttftChan)

	// Calculate maximum and minimum TTFT
	maxTTFT := 0.0
	minTTFT := math.Inf(1)
	validCount := 0
	for ttft := range ttftChan {
		if ttft < 0 {
			// Skip error markers
			continue
		}
		validCount++
		if ttft > maxTTFT {
			maxTTFT = ttft
		}
		if ttft < minTTFT {
			minTTFT = ttft
		}
	}

	// If no valid responses, return default values
	if validCount == 0 {
		log.Printf("Warning: No valid TTFT measurements for concurrency %d", concurrency)
		return 0, 0
	}

	return maxTTFT, minTTFT
}
