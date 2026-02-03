package retriever

import (
	"context"
	"encoding/json"
	"github.com/philippgille/chromem-go"
	"strings"
	"testing"
)

func mockEmbeddingFunc(_ context.Context, text string) ([]float32, error) {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "apple") {
		return []float32{1.0, 0.0, 0.0}, nil
	}
	if strings.Contains(lower, "banana") {
		return []float32{0.0, 1.0, 0.0}, nil
	}
	return []float32{0.0, 0.0, 1.0}, nil
}

func TestCreateLocal(t *testing.T) {
	testData := []KnowledgeEntry{
		{ID: "1", Content: "Apple info", Vector: []float32{1, 0, 0}},
		{ID: "2", Content: "Banana info", Vector: []float32{0, 1, 0}},
		{ID: "3", Content: "Cherry info", Vector: []float32{0, 0, 1}},
	}
	jsonBytes, _ := json.Marshal(testData)

	t.Run("Success Initialization", func(t *testing.T) {
		retriever, err := CreateLocal(mockEmbeddingFunc, 0.7, jsonBytes)
		if err != nil {
			t.Fatalf("CreateLocal failed: %v", err)
		}
		if retriever == nil {
			t.Fatal("Expected retriever instance, got nil")
		}

		res, err := retriever.Search(context.Background(), "apple", 3)
		if err != nil {
			t.Fatalf("Search failed immediately after init: %v", err)
		}
		found := false
		for _, r := range res {
			if strings.Contains(r.Content, "Apple info") {
				found = true
				break
			}
		}
		if !found {
			t.Error("Retriever did not index the initial data correctly")
		}
	})

	t.Run("Invalid JSON", func(t *testing.T) {
		invalidBytes := []byte(`{ not a json }`)
		_, err := CreateLocal(mockEmbeddingFunc, 0.7, invalidBytes)
		if err == nil {
			t.Error("Expected error for invalid JSON, got nil")
		}
	})
}

func TestRetriever_Search(t *testing.T) {
	db := chromem.NewDB()
	coll, err := db.CreateCollection("test_search_coll", nil, mockEmbeddingFunc)
	if err != nil {
		t.Fatalf("Failed to create collection: %v", err)
	}

	docs := []chromem.Document{
		{
			ID:        "1",
			Content:   "Info about Apple",
			Embedding: []float32{1.0, 0.0, 0.0},
		},
		{
			ID:        "2",
			Content:   "Info about Banana",
			Embedding: []float32{0.0, 1.0, 0.0},
		},
		{
			ID:        "3",
			Content:   "Info about Cherry",
			Embedding: []float32{0.0, 0.0, 1.0},
		},
	}

	if err := coll.AddDocuments(context.Background(), docs, 1); err != nil {
		t.Fatalf("Failed to add docs: %v", err)
	}

	retriever := &Retriever{
		coll:          coll,
		minSimilarity: 0.8,
	}

	ctx := context.Background()

	t.Run("High Similarity Match", func(t *testing.T) {

		res, err := retriever.Search(ctx, "apple", 3)
		if err != nil {
			t.Fatalf("Search failed: %v", err)
		}

		foundApple := false
		foundBanana := false
		for _, r := range res {
			if strings.Contains(r.Content, "Info about Apple") {
				foundApple = true
			}
			if strings.Contains(r.Content, "Info about Banana") {
				foundBanana = true
			}
		}
		if !foundApple {
			t.Errorf("Expected 'Info about Apple', got: %v", res)
		}
		if foundBanana {
			t.Error("Should not find Banana (low similarity)")
		}
	})

	t.Run("Empty Query", func(t *testing.T) {
		res, err := retriever.Search(ctx, "", 3)
		if err != nil {
			t.Fatalf("Search error: %v", err)
		}
		if len(res) != 0 {
			t.Error("Expected empty results for empty query")
		}
	})

	t.Run("No Matches (Threshold)", func(t *testing.T) {
		oldSim := retriever.minSimilarity
		retriever.minSimilarity = 2.0
		defer func() { retriever.minSimilarity = oldSim }()

		res, err := retriever.Search(ctx, "apple", 3)
		if err != nil {
			t.Fatalf("Search error: %v", err)
		}
		if len(res) != 0 {
			t.Errorf("Expected no results with similarity > 2.0, got %v", res)
		}
	})
}
