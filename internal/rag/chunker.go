package rag

import "strings"

const (
	DefaultChunkSize = 512  // approximate tokens
	DefaultOverlap   = 50   // token overlap
	CharsPerToken    = 4    // rough approximation
)

// Chunk represents a piece of a document.
type Chunk struct {
	Text  string
	URL   string
	Title string
	Index int
}

// ChunkText splits text into overlapping chunks, breaking at sentence boundaries.
func ChunkText(text, url, title string) []Chunk {
	return ChunkTextWithSize(text, url, title, DefaultChunkSize, DefaultOverlap)
}

func ChunkTextWithSize(text, url, title string, chunkSize, overlap int) []Chunk {
	charChunk := chunkSize * CharsPerToken
	charOverlap := overlap * CharsPerToken

	var chunks []Chunk
	start := 0
	idx := 0

	for start < len(text) {
		end := start + charChunk
		if end > len(text) {
			end = len(text)
		}

		// Try to break at a sentence boundary.
		if end < len(text) {
			for _, delim := range []string{". ", ".\n", "! ", "? ", "?\n"} {
				if pos := strings.LastIndex(text[start:end], delim); pos > 0 {
					end = start + pos + 1
					break
				}
			}
		}

		chunk := strings.TrimSpace(text[start:end])
		if chunk != "" {
			chunks = append(chunks, Chunk{
				Text:  chunk,
				URL:   url,
				Title: title,
				Index: idx,
			})
			idx++
		}

		start = end - charOverlap
		if start < 0 {
			start = 0
		}
		if start >= len(text) {
			break
		}
	}
	return chunks
}
