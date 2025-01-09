package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	hamfts "hamfts/elasticsearch"
)

type SearchRequest struct {
	Query        string `json:"query"`
	ContainsMode bool   `json:"containsMode,omitempty"`
}

type DocumentRequest struct {
	ID      string                 `json:"id"`
	Content string                 `json:"content"`
	Meta    map[string]interface{} `json:"metadata,omitempty"`
}

func main() {
	// Initialize the search index
	idx, err := hamfts.NewIndex("./data")
	if err != nil {
		log.Fatalf("Failed to initialize index: %v", err)
	}
	defer idx.Close()

	// Search endpoint
	http.HandleFunc("/search", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		results, err := idx.Search(req.Query, req.ContainsMode)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(results)
	})

	// Add document endpoint
	http.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var req DocumentRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			doc := hamfts.NewDocument(req.ID, req.Content)
			doc.Metadata = req.Meta

			if err := idx.AddDocument(doc); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)

		case http.MethodGet:
			docs := idx.ListDocumentIDs()
			json.NewEncoder(w).Encode(docs)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Delete document endpoint
	http.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id := r.URL.Path[len("/documents/"):]
		if err := idx.DeleteDocument(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	})

	// Stats endpoint
	http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		stats := idx.GetStats()
		json.NewEncoder(w).Encode(stats)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
