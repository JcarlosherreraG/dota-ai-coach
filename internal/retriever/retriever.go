package retriever

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/philippgille/chromem-go"
	"log"
	"runtime"
	"strings"
)

const (
	storageDir = "../../storage/chroma"
)

// CreateLocal initializes local knowledge storage based on Chroma DB.
// Takes embedding function (BERT), minimum similarity threshold, and JSON knowledge data.
func CreateLocal(embedFunc chromem.EmbeddingFunc, minSimilarity float32, knowledge []byte) (*Retriever, error) {
	log.Println("[Retriever] Initializing Local BERT (Cybertron)...")

	// Create or open an existing persistent DB
	db, err := chromem.NewPersistentDB(storageDir, false)
	if err != nil {
		return nil, fmt.Errorf("failed to create persistent db: %w", err)
	}

	// Get or create a collection for Dota 2 knowledge
	coll, err := db.GetOrCreateCollection("dota2_wiki", nil, embedFunc)
	if err != nil {
		return nil, fmt.Errorf("failed to create collection: %w", err)
	}

	// Decode input knowledge data
	var entries []KnowledgeEntry
	if err := json.Unmarshal(knowledge, &entries); err != nil {
		return nil, fmt.Errorf("corrupted knowledge assets: %w", err)
	}

	// Prepare documents for indexing
	docs := make([]chromem.Document, len(entries))
	for i, entry := range entries {
		docs[i] = chromem.Document{
			ID:        entry.ID,
			Content:   entry.Content,
			Embedding: entry.Vector,
			Metadata:  entry.Metadata,
		}
	}

	// Add documents to the collection in parallel
	if err := coll.AddDocuments(context.Background(), docs, runtime.NumCPU()); err != nil {
		return nil, fmt.Errorf("failed to add docs: %w", err)
	}

	log.Printf("[Retriever] Ready. %d documents indexed.", len(docs))
	return &Retriever{
		coll:          coll,
		minSimilarity: minSimilarity,
	}, nil
}

// Search performs semantic search in the knowledge base.
// Returns a list of results that passed the similarity threshold filter.
func (r *Retriever) Search(ctx context.Context, query string, nResults int) ([]ResultEntry, error) {
	cleanQuery := strings.TrimSpace(query)
	if cleanQuery == "" {
		return nil, nil
	}

	// Perform vector query to the collection
	results, err := r.coll.Query(ctx, query, nResults, nil, nil)
	if err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return nil, nil
	}

	// Filter results by minimum similarity
	var entities []ResultEntry
	for _, res := range results {
		if res.Similarity < r.minSimilarity {
			continue
		}
		entities = append(entities, ResultEntry{
			ID:         res.ID,
			Content:    res.Content,
			Similarity: res.Similarity,
		})
	}
	return entities, nil
}
