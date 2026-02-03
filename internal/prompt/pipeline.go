package prompt

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BrightGir/game-ai-helper/assets"
	"github.com/BrightGir/game-ai-helper/internal/ai"
	"github.com/BrightGir/game-ai-helper/internal/retriever"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	maxRetries    = 3
	retryDelaySec = 2
)

// Pipeline manages the RAG (Retrieval-Augmented Generation) process:
// query generation -> knowledge retrieval -> final prompt construction.
type Pipeline struct {
	ai       ai.Client          // LLM client
	searcher retriever.Searcher // Knowledge base search interface
	tmpl     *Templates         // Prompt templates
}

// NewPipeline creates a new Pipeline instance.
func NewPipeline(aiClient ai.Client, searcher retriever.Searcher) *Pipeline {
	return &Pipeline{
		ai:       aiClient,
		searcher: searcher,
		tmpl:     LoadTemplates(),
	}
}

// Execute runs the full processing cycle: from user question to final prompt with context.
func (p *Pipeline) Execute(ctx context.Context, gameContext string, question string) (string, error) {
	// 1. Generate search queries based on the question and game context
	queries, err := p.generateSearchQueries(ctx, gameContext, question)
	if err != nil {
		log.Printf("[Pipeline] Query generation failed: %v. Using raw user question as fallback.", err)
		queries = []string{question}
	} else {
		log.Printf("[Pipeline] Generated search queries: %v", strings.Join(queries, " | "))
	}

	// 2. Retrieve relevant knowledge from the base
	knowledge, err := p.retrieveKnowledge(ctx, queries)
	if err != nil {
		log.Printf("[Pipeline] RAG failed completely: %v. Proceeding with empty knowledge.", err)
		knowledge = "Knowledge base is temporarily unavailable."
	}
	if knowledge == "" {
		knowledge = "No specific data found in the knowledge base for this situation."
	}

	// 3. Construct the final prompt for the coach
	finalPrompt, err := p.buildFinalPrompt(knowledge, gameContext, question)
	if err != nil {
		return "", fmt.Errorf("failed to build final prompt: %w", err)
	}

	return finalPrompt, nil
}

// generateSearchQueries uses AI to turn a user question into a list of search queries.
func (p *Pipeline) generateSearchQueries(ctx context.Context, gameContext, question string) ([]string, error) {
	var userPromptBuf bytes.Buffer
	// Execute template with data
	err := p.tmpl.QueryGenerator.Execute(&userPromptBuf, map[string]interface{}{
		"GameContext": gameContext,
		"Question":    question,
	})
	if err != nil {
		return nil, fmt.Errorf("template execution failed for query gen: %w", err)
	}
	userPrompt := userPromptBuf.String()

	var lastErr error
	// Retry on errors (e.g., if AI returned invalid JSON)
	for i := 0; i < maxRetries; i++ {
		log.Printf("[Pipeline] Attempt %d/%d to generate search queries...", i+1, maxRetries)

		queries, err := p.executeSingleQueryGenerationAttempt(ctx, userPrompt)
		if err == nil {
			log.Printf("[Pipeline] Successfully generated queries on attempt %d.", i+1)
			return queries, nil
		}

		lastErr = fmt.Errorf("attempt %d failed: %w", i+1, err)
		log.Printf("[Pipeline] Query generation attempt failed: %v", lastErr)

		if i < maxRetries-1 {
			time.Sleep(retryDelaySec * time.Second)
		}
	}

	return nil, fmt.Errorf("all %d attempts to generate search queries failed. Last error: %w", maxRetries, lastErr)
}

// executeSingleQueryGenerationAttempt performs one LLM request and parses JSON with query list.
func (p *Pipeline) executeSingleQueryGenerationAttempt(ctx context.Context, userPrompt string) ([]string, error) {
	rawResp, err := p.ai.Ask(ctx, assets.BaseCoachSystemPrompt, userPrompt)
	if err != nil {
		return nil, fmt.Errorf("llm call failed: %w", err)
	}

	var intent struct {
		Queries []string `json:"queries"`
	}
	// Extract JSON object from text response (in case AI added extra text)
	jsonStr := extractJSONObject(rawResp)
	if jsonStr == "" {
		return nil, fmt.Errorf("could not extract JSON object from LLM response. Raw response: %s", rawResp)
	}

	if err := json.Unmarshal([]byte(jsonStr), &intent); err != nil {
		return nil, fmt.Errorf("failed to parse queries from llm response: %w. JSON string: %s", err, jsonStr)
	}

	if len(intent.Queries) == 0 {
		log.Println("[Pipeline] llm generated an empty query list")
		return []string{}, nil
	}

	return intent.Queries, nil
}

// buildFinalPrompt constructs the final prompt text, combining knowledge, game state, and question.
func (p *Pipeline) buildFinalPrompt(knowledge, gameState, question string) (string, error) {
	var finalPromptBuf bytes.Buffer
	err := p.tmpl.FinalPrompt.Execute(&finalPromptBuf, map[string]interface{}{
		"Knowledge": knowledge,
		"GameState": gameState,
		"Question":  question,
	})
	if err != nil {
		return "", fmt.Errorf("template execution failed for final prompt: %w", err)
	}
	return finalPromptBuf.String(), nil
}

// retrieveKnowledge performs parallel search in the knowledge base for a list of queries.
func (p *Pipeline) retrieveKnowledge(ctx context.Context, queries []string) (string, error) {
	if len(queries) == 0 {
		return "", nil
	}

	uniqueDocs := make(map[string]string)
	var mu sync.Mutex
	var wg sync.WaitGroup
	errs := make(chan error, len(queries))

	for _, q := range queries {
		query := q
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Check if context is cancelled
			if ctx.Err() != nil {
				errs <- fmt.Errorf("search for '%s' aborted due to context cancellation: %w", query, ctx.Err())
				return
			}

			// Perform search (top-10 results)
			results, err := p.searcher.Search(ctx, query, 10)
			if err != nil {
				log.Printf("[Pipeline] WARNING: Search for '%s' failed: %v. Continuing with other queries.", query, err)
				errs <- fmt.Errorf("search for '%s' failed: %w", query, err)
				return
			}

			// Store unique documents
			mu.Lock()
			for _, res := range results {
				uniqueDocs[res.ID] = res.Content
			}
			mu.Unlock()
		}()
	}

	wg.Wait()
	close(errs)

	var allErrs []error
	for err := range errs {
		allErrs = append(allErrs, err)
	}

	// Collect all found fragments into one string
	var resultParts []string
	for _, doc := range uniqueDocs {
		resultParts = append(resultParts, doc)
	}

	knowledge := strings.Join(resultParts, "\n---\n")

	if knowledge != "" {
		if len(allErrs) > 0 {
			log.Printf("[Pipeline] Some search queries failed, but partial knowledge was retrieved. Errors: %v", errors.Join(allErrs...))
		}
		return knowledge, nil
	} else if len(allErrs) > 0 {
		return "", fmt.Errorf("no knowledge could be retrieved; all search queries failed: %w", errors.Join(allErrs...))
	}

	return "", nil
}

// extractJSONObject finds the first and last curly braces and returns the substring between them.
func extractJSONObject(s string) string {
	start := strings.Index(s, "{")
	end := strings.LastIndex(s, "}")
	if start == -1 || end == -1 || end < start {
		return ""
	}
	return s[start : end+1]
}
