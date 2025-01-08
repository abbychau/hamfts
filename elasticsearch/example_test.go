package hamfts

import (
	"fmt"
	"os"
	"testing"
	"time"
)

func TestHamfts(t *testing.T) {
	// Create temporary directory for test
	testDir, err := os.MkdirTemp("", "hamfts_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	// Create new index
	idx, err := NewIndex(testDir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	// Create documents
	doc1 := NewDocument("1", "The quick brown fox jumps over the lazy dog")
	doc1.Metadata["category"] = "animals"
	doc1.CreatedAt = time.Now()

	doc2 := NewDocument("2", "The lazy cat sleeps all day")
	doc2.Metadata["category"] = "animals"
	doc2.CreatedAt = time.Now()

	// Add documents to index
	if err := idx.AddDocument(doc1); err != nil {
		t.Fatal(err)
	}
	if err := idx.AddDocument(doc2); err != nil {
		t.Fatal(err)
	}

	// Search for documents
	results, err := idx.Search("lazy")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	results, err = idx.Search("quick fox")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(results))
	}

	// Get document by ID
	doc, err := idx.GetDocument("1")
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil {
		t.Error("Expected document not found")
	}

	// Delete document
	if err := idx.DeleteDocument("1"); err != nil {
		t.Fatal(err)
	}
	results, err = idx.Search("quick")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 0 {
		t.Errorf("Expected 0 results after deletion, got %d", len(results))
	}
}

func TestDocumentManagement(t *testing.T) {
	testDir, err := os.MkdirTemp("", "hamfts_test_docs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	idx, err := NewIndex(testDir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	// Test document count
	if count := idx.DocumentCount(); count != 0 {
		t.Errorf("Expected initial count 0, got %d", count)
	}

	// Add documents
	docs := []*Document{
		NewDocument("1", "first document"),
		NewDocument("2", "second document"),
		NewDocument("3", "third document"),
	}

	for _, doc := range docs {
		if err := idx.AddDocument(doc); err != nil {
			t.Fatal(err)
		}
	}

	// Test document count after additions
	if count := idx.DocumentCount(); count != 3 {
		t.Errorf("Expected count 3, got %d", count)
	}

	// Test document retrieval
	doc, err := idx.GetDocument("2")
	if err != nil {
		t.Fatal(err)
	}
	if doc == nil || doc.ID != "2" {
		t.Error("Failed to retrieve correct document")
	}

	// Test document listing
	ids := idx.ListDocumentIDs()
	if len(ids) != 3 {
		t.Errorf("Expected 3 document IDs, got %d", len(ids))
	}

	// Test deletion
	if err := idx.DeleteDocument("2"); err != nil {
		t.Fatal(err)
	}

	// Verify deletion
	if count := idx.DocumentCount(); count != 2 {
		t.Errorf("Expected count 2 after deletion, got %d", count)
	}

	doc, err = idx.GetDocument("2")
	if err != nil {
		t.Fatal(err)
	}
	if doc != nil {
		t.Error("Document should have been deleted")
	}
}

func TestBulkOperations(t *testing.T) {
	testDir, err := os.MkdirTemp("", "hamfts_test_bulk")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(testDir)

	idx, err := NewIndex(testDir)
	if err != nil {
		t.Fatal(err)
	}
	defer idx.Close()

	// Create test documents
	docs := make([]*Document, 100)
	for i := 0; i < 100; i++ {
		docs[i] = NewDocument(
			fmt.Sprintf("doc%d", i),
			fmt.Sprintf("content for document %d with some common words", i),
		)
	}

	// Test bulk addition
	if err := idx.AddDocuments(docs); err != nil {
		t.Fatal(err)
	}

	// Test search after bulk addition
	results, err := idx.Search("common words")
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 100 {
		t.Errorf("Expected 100 results, got %d", len(results))
	}

	// Test compaction
	for i := 0; i < 50; i++ {
		if err := idx.DeleteDocument(fmt.Sprintf("doc%d", i)); err != nil {
			t.Fatal(err)
		}
	}

	if err := idx.Compact(); err != nil {
		t.Fatal(err)
	}

	// Verify after compaction
	stats := idx.GetStats()
	if stats["documentCount"] != 50 {
		t.Errorf("Expected 50 documents after compaction, got %d", stats["documentCount"])
	}
}
