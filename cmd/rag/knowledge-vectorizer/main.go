// Package main implements a tool to vectorize knowledge chunks using BERT and store them in a JSON database for RAG.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/BrightGir/game-ai-helper/cmd/rag"
	"github.com/BrightGir/game-ai-helper/internal/retriever"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	SourcesDir = "assets/data"           // Directory containing processed JSON chunks
	OutputFile = "assets/rag/knowledge.json" // Final RAG database file
	MaxWorkers = 10                      // Parallel workers for embedding generation
)

func main() {
	start := time.Now()
	log.Println("STARTING KNOWLEDGE VECTORIZER")

	numWorkers := MaxWorkers
	log.Printf("Loading BERT Model... Will use %d parallel workers.", numWorkers)

	// Initialize BERT encoder function
	encoder, err := retriever.CreateEmbedFunction()
	if err != nil {
		log.Fatalf("Model initialization failed: %v", err)
	}

	// Scan for JSON files in the sources directory
	files, err := getJsonFiles(SourcesDir)
	if err != nil {
		log.Fatalf("Scan failed: %v", err)
	}
	log.Printf("Found %d JSON files to process.", len(files))

	type Job struct {
		Filename string
		Index    int
		Chunk    rag.KnowledgeChunk
	}
	var jobs []Job

	// 1. Extract all chunks from files and prepare jobs
	for _, file := range files {
		contentBytes, err := os.ReadFile(file)
		if err != nil {
			log.Printf("Error reading file %s: %v", file, err)
			continue
		}

		var fileChunks []rag.KnowledgeChunk
		if err := json.Unmarshal(contentBytes, &fileChunks); err != nil {
			log.Printf("Error parsing JSON in %s: %v", file, err)
			continue
		}

		filename := filepath.Base(file)
		for i, chunk := range fileChunks {
			if strings.TrimSpace(chunk.SearchText) == "" {
				continue
			}
			jobs = append(jobs, Job{Filename: filename, Index: i, Chunk: chunk})
		}
	}

	totalJobs := len(jobs)
	log.Printf("Extracted %d chunks. Starting vectorization...", totalJobs)

	var entries []retriever.KnowledgeEntry
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit parallel embedding requests
	sem := make(chan struct{}, numWorkers)
	processed := 0

	// 2. Process jobs in parallel
	for _, job := range jobs {
		wg.Add(1)

		go func(j Job) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire token
			defer func() { <-sem }() // Release token

			cleanTags := strings.Join(j.Chunk.Tags, ", ")
			searchableText := j.Chunk.SearchText

			// Prepend tags to search text for better semantic matching
			var tagPrefix string
			if len(j.Chunk.Tags) > 0 {
				tagPrefix = fmt.Sprintf("[TAGS: %s] | ", cleanTags)
				searchableText = tagPrefix + searchableText
			}

			// Generate embedding vector for the chunk
			vector, err := encoder(context.Background(), searchableText)
			if err != nil {
				log.Printf("Embedding error %s[%d]: %v", j.Filename, j.Index, err)
				return
			}

			// Create final knowledge entry
			entry := retriever.KnowledgeEntry{
				ID:      fmt.Sprintf("%s_%d", j.Filename, j.Index),
				Content: tagPrefix + j.Chunk.DisplayText,
				Vector:  vector,
			}
			
			mu.Lock()
			entries = append(entries, entry)
			processed++
			fmt.Printf("\rProgress: %d / %d", processed, totalJobs)
			mu.Unlock()
		}(job)
	}

	wg.Wait()
	fmt.Println()

	// 3. Save all vectorized entries to the final JSON file
	saveJSON(entries)
	log.Printf("DONE. Database built with %d entries in %v", len(entries), time.Since(start))
}

// saveJSON writes vectorized entries to the output file.
func saveJSON(entries []retriever.KnowledgeEntry) {
	_ = os.MkdirAll(filepath.Dir(OutputFile), 0755)
	f, err := os.Create(OutputFile)
	if err != nil {
		log.Fatalf("Failed to create output file: %v", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(entries); err != nil {
		log.Fatal(err)
	}
}

// getJsonFiles returns a list of all .json files in a directory.
func getJsonFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".json" {
				files = append(files, path)
			}
		}
		return nil
	})
	return files, err
}
