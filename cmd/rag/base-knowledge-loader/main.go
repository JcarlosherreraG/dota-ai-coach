// Package main implements a tool to fetch raw Dota 2 data and process it into knowledge chunks using LLM.
package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/BrightGir/game-ai-helper/assets"
	"github.com/BrightGir/game-ai-helper/cmd/rag"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/ai/openrouter"
	"github.com/joho/godotenv"
)

const (
	CacheDir      = "storage/cache"          // Directory for caching LLM responses
	OutputDataDir = "assets/data"            // Directory for processed JSON data
	MaxWorkers    = 20                       // Number of parallel workers for LLM processing
	Model         = "deepseek/deepseek-v3.2" // LLM model to use for chunking
)

// SourceTask defines a data source and how it should be processed.
type SourceTask struct {
	Name          string                                           // Task name (e.g., "abilities")
	URL           string                                           // Url to fetch
	FetchFunction func(url string, token string) ([]string, error) // Fetch function returns batches
	SystemPrompt  string                                           // Instruction for the LLM on how to process this data
	BatchSize     int                                              // Number of objects to send to LLM in one request
	token         string                                           // Auth token for fetch
}

// KnowledgeProcessor handles the end-to-end processing of a knowledge source.
type KnowledgeProcessor struct {
	Client ai.Client // AI client for processing data
}

func main() {
	setupEnv()
	apiKey := os.Getenv("KNOWLEDGE_LOADER_AI_API_KEY")
	if apiKey == "" {
		log.Fatal("ERROR: variable KNOWLEDGE_LOADER_AI_API_KEY not set.")
	}
	stratzApiKey := os.Getenv("STRATZ_API_KEY")
	if stratzApiKey == "" {
		log.Fatal("ERROR: variable STRATZ_API_KEY not set.")
	}

	// Initialize OpenRouter client for data processing
	client := openrouter.NewClient(apiKey, Model, 300)
	processor := &KnowledgeProcessor{Client: client}

	// Define tasks for different Dota 2 data sources
	tasks := []SourceTask{
		{
			Name: "items",
			URL:  "https://raw.githubusercontent.com/odota/dotaconstants/master/build/items.json",

			FetchFunction: fetchAndSplitJSON,
			SystemPrompt:  assets.ItemsProcessingPrompt,
			BatchSize:     1,
		},
		{
			Name:          "heroes",
			URL:           "https://api.stratz.com/graphql",
			token:         stratzApiKey,
			FetchFunction: fetchAndEnrichStratz,
			SystemPrompt:  assets.HeroesProcessingPrompt,
			BatchSize:     1,
		},
		{
			Name:          "aghs_desc",
			URL:           "https://raw.githubusercontent.com/odota/dotaconstants/master/build/aghs_desc.json",
			FetchFunction: fetchAndSplitJSON,
			SystemPrompt:  assets.AganimProcessingPrompt,
			BatchSize:     1,
		},
	}

	// Ensure output directory exists
	if err := os.MkdirAll(OutputDataDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Run each task sequentially
	for _, task := range tasks {
		processor.Run(task)
	}
}

// Run executes a single source processing task.
func (kp *KnowledgeProcessor) Run(task SourceTask) {
	log.Printf("Starting process source: %s", task.Name)

	// 1. Fetch raw data and split into individual objects
	rawObjects, err := task.FetchFunction(task.URL, task.token)
	if err != nil {
		log.Printf("ERROR fetching %s: %v", task.Name, err)
		return
	}

	// 2. Setup cache directory
	cacheDir := filepath.Join(CacheDir, task.Name)
	os.MkdirAll(cacheDir, 0755)

	// 3. Create batches for LLM processing
	batches := kp.createBatches(task.BatchSize, rawObjects)
	totalBatches := len(batches)

	var (
		allChunks      []rag.KnowledgeChunk
		mu             sync.Mutex
		wg             sync.WaitGroup
		processedCount int32
	)

	jobsChan := make(chan []string, totalBatches)

	// 4. Start worker pool for parallel processing
	for i := 0; i < MaxWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for batch := range jobsChan {
				chunks, isCache, err := kp.processBatch(batch, task, cacheDir)
				if err != nil {
					log.Printf("Batch processing error: %v", err)
					continue
				}

				mu.Lock()
				allChunks = append(allChunks, chunks...)
				mu.Unlock()

				current := atomic.AddInt32(&processedCount, 1)
				if isCache {
					log.Printf("[%s] Progress: %d/%d batches (Cache Hit)", task.Name, current, totalBatches)
				} else {
					log.Printf("[%s] Progress: %d/%d batches", task.Name, current, totalBatches)
				}
			}
		}()
	}

	// 5. Feed batches into the worker pool
	for _, b := range batches {
		jobsChan <- b
	}
	close(jobsChan)
	wg.Wait()

	// 6. Save the aggregated result to a file
	kp.saveFinalResult(task.Name, allChunks)
}

// processBatch handles a single batch of data, checking cache first.
func (kp *KnowledgeProcessor) processBatch(batch []string, task SourceTask, cacheDir string) ([]rag.KnowledgeChunk, bool, error) {
	batchRawStr := strings.Join(batch, "")

	// Calculate hash for caching
	hashInput := batchRawStr
	batchHash := fmt.Sprintf("%x", md5.Sum([]byte(hashInput)))
	cachePath := filepath.Join(cacheDir, batchHash+".json")

	var results []rag.KnowledgeChunk

	// Try loading from cache
	if data, err := os.ReadFile(cachePath); err == nil {
		if json.Unmarshal(data, &results) == nil {
			return results, true, nil
		}
	}

	// If not in cache, call LLM to process the batch
	batchJsonPayload := "[" + strings.Join(batch, ",") + "]"
	results, err := kp.callLLM(task.Name, batchJsonPayload, task.SystemPrompt)
	if err != nil {
		return nil, false, err
	}

	// Save to cache for future runs
	kp.saveJSON(cachePath, results)
	return results, false, nil
}

// callLLM sends data to the LLM and parses the resulting knowledge chunks.
func (kp *KnowledgeProcessor) callLLM(contextName, rawJson, systemPrompt string) ([]rag.KnowledgeChunk, error) {
	userPrompt := fmt.Sprintf("SOURCE TYPE: %s\n\nRAW DATA BATCH:\n%s", contextName, rawJson)
	maxRetries := 3

	for i := 0; i < maxRetries; i++ {
		responseStr, err := kp.Client.Ask(context.Background(), userPrompt, systemPrompt)
		if err != nil {
			log.Printf("LLM Ask failed (attempt %d): %v", i+1, err)
			kp.retryPause(i)
			continue
		}

		// Clean markdown formatting if present
		cleanedJson := kp.cleanMarkdown(responseStr)
		var result []rag.KnowledgeChunk
		if err := json.Unmarshal([]byte(cleanedJson), &result); err != nil {
			log.Printf("JSON unmarshal error: %v. Raw: %s", err, responseStr)
			kp.retryPause(i)
			continue
		}
		return result, nil
	}
	return nil, fmt.Errorf("failed after retries")
}

// createBatches splits a list of raw objects into smaller batches.
func (kp *KnowledgeProcessor) createBatches(batchSize int, rawObjects []string) [][]string {
	var batches [][]string
	for i := 0; i < len(rawObjects); i += batchSize {
		end := i + batchSize
		if end > len(rawObjects) {
			end = len(rawObjects)
		}

		batch := make([]string, end-i)
		copy(batch, rawObjects[i:end])
		batches = append(batches, batch)
	}
	return batches
}

// saveFinalResult writes the processed chunks to the final assets directory.
func (kp *KnowledgeProcessor) saveFinalResult(sourceName string, chunks []rag.KnowledgeChunk) {
	saveFile := filepath.Join(OutputDataDir, sourceName+".json")
	kp.saveJSON(saveFile, chunks)
	log.Printf("Saved %d chunks to %s", len(chunks), saveFile)
}

// cleanMarkdown removes markdown code block wrappers from LLM response.
func (kp *KnowledgeProcessor) cleanMarkdown(input string) string {
	start := strings.Index(input, "[")
	end := strings.LastIndex(input, "]")

	if start == -1 || end == -1 || start > end {
		return strings.TrimSpace(input)
	}

	return input[start : end+1]
}

// retryPause implements exponential backoff for retries.
func (kp *KnowledgeProcessor) retryPause(attempt int) {
	pauseTime := time.Duration(math.Pow(2, float64(attempt))) * time.Second
	log.Printf("Retry in %v...", pauseTime)
	time.Sleep(pauseTime)
}

// saveJSON writes data to a file in indented JSON format.
func (kp *KnowledgeProcessor) saveJSON(path string, data interface{}) {
	f, _ := os.Create(path)
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	enc.Encode(data)
}

// setupEnv loads environment variables from .env file.
func setupEnv() {
	if err := godotenv.Load(); err != nil {
		log.Println("WARN: .env not found, will use system variables")
	}
}
