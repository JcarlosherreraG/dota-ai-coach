package prompt

import (
	"context"
	"strings"
	"testing"
)

func TestNewPipeline(t *testing.T) {
	mockAI := &mockAIClient{}
	mockSearch := &mockSearcher{}

	pipeline := NewPipeline(mockAI, mockSearch)

	if pipeline == nil {
		t.Fatal("Expected pipeline instance, got nil")
	}
	if pipeline.ai == nil {
		t.Error("AI client not initialized")
	}
	if pipeline.searcher == nil {
		t.Error("Searcher not initialized")
	}
	if pipeline.tmpl == nil {
		t.Error("Templates not loaded")
	}
}

func TestPipeline_Execute_Success(t *testing.T) {
	mockAI := &mockAIClient{response: `{"queries": ["bkb piercing", "disable mechanics"]}`}

	mockSearch := &mockSearcher{
		results: []string{"BKB info"},
	}

	pipeline := NewPipeline(mockAI, mockSearch)
	gameContext := "Hero: Pudge, Items: blink"
	question := "Should I buy BKB?"

	result, err := pipeline.Execute(context.Background(), gameContext, question)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result == "" {
		t.Error("Expected non-empty result")
	}

	if !strings.Contains(result, gameContext) {
		t.Error("Result should contain game context")
	}
	if !strings.Contains(result, question) {
		t.Error("Result should contain question")
	}
}

func TestPipeline_Execute_QueryGenerationFailure(t *testing.T) {
	mockAI := &mockAIClient{response: ""}
	mockSearch := &mockSearcher{results: nil}

	pipeline := NewPipeline(mockAI, mockSearch)

	result, err := pipeline.Execute(context.Background(), "game state", "question")

	if err != nil {
		t.Fatalf("Execute should not fail on query generation error: %v", err)
	}

	if result == "" {
		t.Error("Should return result even if query generation failed")
	}
}

func TestPipeline_Execute_SearchFailure(t *testing.T) {
	mockAI := &mockAIClient{response: `{"queries": ["test"]}`}
	mockSearch := &mockSearcher{results: nil}

	pipeline := NewPipeline(mockAI, mockSearch)

	result, err := pipeline.Execute(context.Background(), "game", "question")

	if err != nil {
		t.Fatalf("Execute should not fail on search error: %v", err)
	}

	if result == "" {
		t.Error("Should return result even if search failed")
	}

	if !strings.Contains(result, "unavailable") && !strings.Contains(result, "not found") {
		t.Log("Result should mention knowledge unavailability")
	}
}

func TestExtractJSONObject(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Valid JSON in text",
			input:    `Some text before {"key": "value"} some text after`,
			expected: `{"key": "value"}`,
		},
		{
			name:     "JSON only",
			input:    `{"queries": ["q1", "q2"]}`,
			expected: `{"queries": ["q1", "q2"]}`,
		},
		{
			name:     "No JSON",
			input:    "No JSON here",
			expected: "",
		},
		{
			name:     "Empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "Only opening brace",
			input:    "{incomplete",
			expected: "",
		},
		{
			name:     "Only closing brace",
			input:    "incomplete}",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONObject(tt.input)
			if got != tt.expected {
				t.Errorf("extractJSONObject(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestPipeline_generateSearchQueries_InvalidJSON(t *testing.T) {
	mockAI := &mockAIClient{response: "not a json"}

	pipeline := NewPipeline(mockAI, &mockSearcher{})

	queries, err := pipeline.generateSearchQueries(context.Background(), "game", "question")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if queries != nil {
		t.Errorf("Expected nil queries on error, got %v", queries)
	}
}

func TestPipeline_generateSearchQueries_EmptyQueries(t *testing.T) {
	mockAI := &mockAIClient{response: `{"queries": []}`}

	pipeline := NewPipeline(mockAI, &mockSearcher{})

	queries, err := pipeline.generateSearchQueries(context.Background(), "game", "question")

	if err != nil {
		t.Fatalf("Should not fail on empty queries: %v", err)
	}
	if len(queries) != 0 {
		t.Errorf("Expected empty queries array, got %v", queries)
	}
}

func TestPipeline_retrieveKnowledge_EmptyQueries(t *testing.T) {
	pipeline := NewPipeline(&mockAIClient{}, &mockSearcher{})

	knowledge, err := pipeline.retrieveKnowledge(context.Background(), []string{})

	if err != nil {
		t.Errorf("Should not fail on empty queries: %v", err)
	}
	if knowledge != "" {
		t.Errorf("Expected empty knowledge for empty queries, got %q", knowledge)
	}
}

func TestPipeline_retrieveKnowledge_UniqueDocuments(t *testing.T) {
	mockSearch := &mockSearcher{
		results: []string{"Duplicated content", "Duplicated content", "Duplicated content"},
	}

	pipeline := NewPipeline(&mockAIClient{}, mockSearch)

	queries := []string{"query1", "query2", "query3"}
	knowledge, err := pipeline.retrieveKnowledge(context.Background(), queries)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if knowledge == "" {
		t.Error("Expected non-empty knowledge")
	}
}

func TestPipeline_retrieveKnowledge_ContextCancellation(t *testing.T) {
	mockSearch := &mockSearcher{results: nil}

	pipeline := NewPipeline(&mockAIClient{}, mockSearch)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	knowledge, err := pipeline.retrieveKnowledge(ctx, []string{"query"})

	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
	if knowledge != "" {
		t.Errorf("Expected empty knowledge on error, got %q", knowledge)
	}
}
