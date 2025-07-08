package utils

import (
	"log"
	"math/rand"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Yoosu-L/llmapibenchmark/internal/api"

	"github.com/sashabaranov/go-openai"
)

// MeasureSpeed measures API generation throughput and TTFT.
func MeasureSpeed(baseURL, apiKey, model, prompt string, concurrency, maxTokens int, latency float64) (float64, float64, float64, float64) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	var wg sync.WaitGroup
	var responseTokens sync.Map
	var promptTokens sync.Map
	var successCount int32 // ← 添加成功计数
	var errorCount int32   // ← 添加错误计数

	// Measure TTFT
	maxTTFT, minTTFT := api.MeasureTTFT(client, model, prompt, concurrency)

	start := time.Now()

	// Send requests concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			resp, err := api.AskOpenAI(client, model, prompt, maxTokens)
			if err != nil {
				log.Printf("Request %d failed: %v", index, err)
				atomic.AddInt32(&errorCount, 1) // ← 记录错误
				responseTokens.Store(index, 0)  // ← 存储0值而不是跳过
				promptTokens.Store(index, 0)
				return
			}
			atomic.AddInt32(&successCount, 1) // ← 记录成功
			responseTokens.Store(index, resp.Usage.CompletionTokens)
			promptTokens.Store(index, resp.Usage.PromptTokens)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate total tokens
	totalResponseTokens := 0
	responseTokens.Range(func(_, value interface{}) bool {
		totalResponseTokens += value.(int)
		return true
	})

	totalPromptTokens := 0
	promptTokens.Range(func(_, value interface{}) bool {
		totalPromptTokens += value.(int)
		return true
	})
	generationSpeed := float64(totalResponseTokens) / (duration.Seconds() - latency/1000)

	// Calculate Prompt Throughput
	promptThroughput := float64(totalPromptTokens) / (maxTTFT - latency/1000)

	// Log statistics - 移动到这里
	successCountValue := atomic.LoadInt32(&successCount)
	errorCountValue := atomic.LoadInt32(&errorCount)

	// 只有在有失败时才打印错误信息
	if errorCountValue > 0 {
		log.Printf("Concurrency %d: %d successful, %d failed requests",
			concurrency, successCountValue, errorCountValue)

		if errorCountValue > int32(concurrency/2) {
			log.Printf("Warning: More than 50%% requests failed for concurrency %d", concurrency)
		}
	}

	return generationSpeed, promptThroughput, maxTTFT, minTTFT
}

const (
	minWordLength = 3
	maxWordLength = 10
)

var letters = []rune("abcdefghijklmnopqrstuvwxyz")

func generateRandomWord() string {
	// length（3-10）
	wordLength := minWordLength + rand.Intn(maxWordLength-minWordLength+1)

	word := make([]rune, wordLength)

	for i := 0; i < wordLength; i++ {
		word[i] = letters[rand.Intn(len(letters))]
	}

	return string(word)
}

func generateRandomPhrase(numWords int) string {
	rand.Seed(time.Now().UnixNano())

	randomWords := make([]string, numWords)
	for i := 0; i < numWords; i++ {
		randomWords[i] = generateRandomWord()
	}

	randomPhrase := strings.Join(randomWords, " ")

	result := "Please reply back the following section unchanged: " + randomPhrase

	return result
}

func MeasureSpeedwithRandomInput(baseURL, apiKey, model string, numWords int, concurrency, maxTokens int, latency float64) (float64, float64, float64, float64) {
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = baseURL
	client := openai.NewClientWithConfig(config)

	var wg sync.WaitGroup
	var responseTokens sync.Map
	var promptTokens sync.Map
	var successCount int32 // ← 添加成功计数
	var errorCount int32   // ← 添加错误计数

	// Measure TTFT
	maxTTFT, minTTFT := api.MeasureTTFTwithRandomInput(client, model, numWords, concurrency)

	start := time.Now()

	// Send requests concurrently
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			prompt := generateRandomPhrase(numWords)
			resp, err := api.AskOpenAI(client, model, prompt, maxTokens)
			if err != nil {
				log.Printf("Request %d failed: %v", index, err)
				atomic.AddInt32(&errorCount, 1) // ← 记录错误
				responseTokens.Store(index, 0)  // ← 存储0值而不是跳过
				promptTokens.Store(index, 0)
				return
			}
			atomic.AddInt32(&successCount, 1) // ← 记录成功
			responseTokens.Store(index, resp.Usage.CompletionTokens)
			promptTokens.Store(index, resp.Usage.PromptTokens)
		}(i)
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate total tokens
	totalResponseTokens := 0
	responseTokens.Range(func(_, value interface{}) bool {
		totalResponseTokens += value.(int)
		return true
	})

	totalPromptTokens := 0
	promptTokens.Range(func(_, value interface{}) bool {
		totalPromptTokens += value.(int)
		return true
	})

	// Calculate speed (tokens/second)
	generationSpeed := float64(totalResponseTokens) / duration.Seconds()

	// Prompt Throughput: 输入token数 / TTFT时间（因为TTFT就是处理输入的时间）
	promptThroughput := float64(totalPromptTokens) / maxTTFT

	successCountValue := atomic.LoadInt32(&successCount)
	errorCountValue := atomic.LoadInt32(&errorCount)

	// 只有在有失败时才打印错误信息
	if errorCountValue > 0 {
		log.Printf("Concurrency %d: %d successful, %d failed requests",
			concurrency, successCountValue, errorCountValue)

		if errorCountValue > int32(concurrency/2) {
			log.Printf("Warning: More than 50%% requests failed for concurrency %d", concurrency)
		}
	}

	return generationSpeed, promptThroughput, maxTTFT, minTTFT
}
