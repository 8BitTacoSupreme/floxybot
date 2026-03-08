package main

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
)

func canonHandler(canonDir string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snapPath := filepath.Join(canonDir, "canon.gob")

		info, err := os.Stat(snapPath)
		if err != nil {
			http.Error(w, "canon snapshot not found", http.StatusNotFound)
			return
		}

		// Compute ETag from file size + mod time.
		etag := fmt.Sprintf(`"%x"`, sha256.Sum256([]byte(fmt.Sprintf("%d-%d", info.Size(), info.ModTime().UnixNano()))))

		// Check If-None-Match for conditional request.
		if r.Header.Get("If-None-Match") == etag {
			w.WriteHeader(http.StatusNotModified)
			return
		}

		w.Header().Set("ETag", etag)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=canon.gob")
		http.ServeFile(w, r, snapPath)
	}
}
