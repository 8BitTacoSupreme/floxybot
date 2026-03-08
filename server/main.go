package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	canonDir := os.Getenv("CANON_DIR")
	if canonDir == "" {
		canonDir = "./data/canon"
	}

	dbURL := os.Getenv("DATABASE_URL")

	var store FeedbackStore
	if dbURL != "" {
		var err error
		store, err = NewPostgresStore(dbURL)
		if err != nil {
			log.Fatalf("connecting to database: %v", err)
		}
		defer store.Close()
		log.Println("Using PostgreSQL for feedback storage")
	} else {
		store = &NoopStore{}
		log.Println("Warning: DATABASE_URL not set, feedback will not be persisted")
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /canon/latest", canonHandler(canonDir))
	mux.HandleFunc("POST /feedback", feedbackHandler(store))
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})

	addr := fmt.Sprintf(":%s", port)
	log.Printf("Floxybot backend listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}
