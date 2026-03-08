package canon

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/8BitTacoSupreme/floxybot/internal/rag"
)

var baseURLs = []string{
	"https://flox.dev/docs/",
	"https://flox.dev/blog/",
}

const defaultMaxPages = 50

// BuildCanon scrapes flox.dev in batches, saving after each batch.
// If it crashes partway through, the last saved snapshot is still usable.
func BuildCanon(ctx context.Context, outputPath string) error {
	// Load existing snapshot if present (incremental build).
	var allChunks []rag.Chunk
	if existing, err := LoadSnapshot(outputPath); err == nil {
		allChunks = existing.Chunks
		fmt.Printf("Loaded existing snapshot with %d chunks\n", len(allChunks))
	}

	scraper := NewScraper()
	totalPages := 0
	batchSize := 10

	for _, base := range baseURLs {
		fmt.Printf("Crawling %s ...\n", base)
		var batch []rag.Chunk
		batchPages := 0

		scraper.Crawl(base, defaultMaxPages, func(p Page) {
			chunks := rag.ChunkText(p.Content, p.URL, p.Title)
			batch = append(batch, chunks...)
			totalPages++
			batchPages++

			// Save every batchSize pages.
			if batchPages >= batchSize {
				allChunks = append(allChunks, batch...)
				batch = nil
				batchPages = 0
				if err := saveSnap(outputPath, allChunks); err != nil {
					fmt.Printf("  warning: save failed: %v\n", err)
				} else {
					fmt.Printf("  saved: %d pages, %d chunks\n", totalPages, len(allChunks))
				}
			}
		})

		// Save any remaining batch.
		if len(batch) > 0 {
			allChunks = append(allChunks, batch...)
			if err := saveSnap(outputPath, allChunks); err != nil {
				fmt.Printf("  warning: save failed: %v\n", err)
			} else {
				fmt.Printf("  saved: %d pages, %d chunks\n", totalPages, len(allChunks))
			}
		}
	}

	fmt.Printf("Done: %d chunks from %d pages -> %s\n", len(allChunks), totalPages, outputPath)
	return nil
}

func saveSnap(path string, chunks []rag.Chunk) error {
	os.MkdirAll("data/canon", 0o755)
	snap := &Snapshot{
		Version:   "1",
		CreatedAt: time.Now(),
		Chunks:    chunks,
	}
	return SaveSnapshot(path, snap)
}
