package canon

import (
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/8BitTacoSupreme/floxybot/internal/rag"
)

// Snapshot is the serialized canon: all chunks with their pre-computed embeddings.
type Snapshot struct {
	Version   string
	CreatedAt time.Time
	Chunks    []rag.Chunk
}

// SaveSnapshot writes a snapshot to a gob file.
func SaveSnapshot(path string, snap *Snapshot) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating snapshot file: %w", err)
	}
	defer f.Close()
	return gob.NewEncoder(f).Encode(snap)
}

// LoadSnapshot reads a snapshot from a gob file.
func LoadSnapshot(path string) (*Snapshot, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening snapshot: %w", err)
	}
	defer f.Close()
	var snap Snapshot
	if err := gob.NewDecoder(f).Decode(&snap); err != nil {
		return nil, fmt.Errorf("decoding snapshot: %w", err)
	}
	return &snap, nil
}
