package retriever

import (
	"context"
	"github.com/philippgille/chromem-go"
)

type Searcher interface {
	Search(ctx context.Context, query string, nResults int) ([]ResultEntry, error)
}

type Retriever struct {
	coll          *chromem.Collection
	minSimilarity float32
}

type KnowledgeEntry struct {
	ID       string            `json:"id"`
	Content  string            `json:"text"`
	Vector   []float32         `json:"vector"`
	Metadata map[string]string `json:"metadata"`
}

type ResultEntry struct {
	ID         string
	Content    string
	Similarity float32
}
