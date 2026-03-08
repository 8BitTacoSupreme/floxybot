package main

import (
	"encoding/json"
	"net/http"
)

type FeedbackPayload struct {
	Query        string   `json:"query"`
	ResponseHash string   `json:"response_hash"`
	Vote         string   `json:"vote"`
	Comment      string   `json:"comment,omitempty"`
	Sources      []string `json:"sources,omitempty"`
}

func feedbackHandler(store FeedbackStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload FeedbackPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Query == "" || payload.Vote == "" {
			http.Error(w, "query and vote are required", http.StatusBadRequest)
			return
		}

		if payload.Vote != "up" && payload.Vote != "down" {
			http.Error(w, "vote must be 'up' or 'down'", http.StatusBadRequest)
			return
		}

		if err := store.SaveFeedback(r.Context(), payload); err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}
}
