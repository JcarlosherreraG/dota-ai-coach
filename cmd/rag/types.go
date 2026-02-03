// Package rag defines shared types for RAG-related tools.
package rag

// KnowledgeChunk represents a processed piece of information stored in the knowledge base.
type KnowledgeChunk struct {
	DisplayText string   `json:"display_text"` // Text to be shown in the AI prompt
	SearchText  string   `json:"search_text"`  // Text used for semantic search (vectorization)
	Tags        []string `json:"tags"`         // Metadata tags for better search relevance
}
