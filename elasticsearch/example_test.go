package hamfts

import (
	"testing"
	"time"
)

func TestHamfts(t *testing.T) {
	// Create new index
	idx := NewIndex()

	// Create documents
	doc1 := NewDocument("1", "The quick brown fox jumps over the lazy dog")
	doc1.Metadata["category"] = "animals"
	doc1.CreatedAt = time.Now()

	doc2 := NewDocument("2", "The lazy cat sleeps all day")
	doc2.Metadata["category"] = "animals"
	doc2.CreatedAt = time.Now()

	// Add documents to index
	idx.AddDocument(doc1)
	idx.AddDocument(doc2)

	// Search for documents
	results := idx.Search("lazy")
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	results = idx.Search("quick fox")
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Get document by ID
	doc := idx.GetDocument("1")
	if doc == nil {
		t.Error("Expected document not found")
	}

	// Delete document
	idx.DeleteDocument("1")
	results = idx.Search("quick")
	if len(results) != 0 {
		t.Errorf("Expected 0 results after deletion, got %d", len(results))
	}
}
