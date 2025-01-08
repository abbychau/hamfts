package hamfts

import (
	"strings"
	"sync"
)

type Index struct {
	mutex     sync.RWMutex
	documents map[string]*Document
	inverted  map[string]map[string]struct{}
}

func NewIndex() *Index {
	return &Index{
		documents: make(map[string]*Document),
		inverted:  make(map[string]map[string]struct{}),
	}
}

func (idx *Index) AddDocument(doc *Document) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Store the document
	idx.documents[doc.ID] = doc

	// Create inverted index
	words := strings.Fields(strings.ToLower(doc.Content))
	for _, word := range words {
		// trim punctuation
		word = strings.Trim(word, ",.!? \t\n\r")

		if idx.inverted[word] == nil {
			idx.inverted[word] = make(map[string]struct{})
		}
		idx.inverted[word][doc.ID] = struct{}{}
	}
}

func (idx *Index) Search(query string) []*Document {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil
	}

	// Get document IDs containing the first word
	results := make(map[string]struct{})
	for docID := range idx.inverted[words[0]] {
		results[docID] = struct{}{}
	}

	// Intersect with other words
	for _, word := range words[1:] {
		temp := make(map[string]struct{})
		for docID := range idx.inverted[word] {
			if _, ok := results[docID]; ok {
				temp[docID] = struct{}{}
			}
		}
		results = temp
	}

	// Convert results to slice of documents
	docs := make([]*Document, 0, len(results))
	for docID := range results {
		docs = append(docs, idx.documents[docID])
	}
	return docs
}

func (idx *Index) GetDocument(id string) *Document {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return idx.documents[id]
}

func (idx *Index) DeleteDocument(id string) {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	doc, exists := idx.documents[id]
	if !exists {
		return
	}

	// Remove from inverted index
	words := strings.Fields(strings.ToLower(doc.Content))
	for _, word := range words {
		delete(idx.inverted[word], id)
		if len(idx.inverted[word]) == 0 {
			delete(idx.inverted, word)
		}
	}

	// Remove document
	delete(idx.documents, id)
}
