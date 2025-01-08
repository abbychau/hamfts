package hamfts

import (
	"encoding/gob"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type IndexMetadata struct {
	DocumentCount     int
	IndexEntries      map[string][]int64 // word -> file positions
	DocumentPositions map[string]int64   // docID -> file position
}

type Index struct {
	mutex     sync.RWMutex
	baseDir   string
	metadata  IndexMetadata
	docFile   *os.File
	indexFile *os.File
}

func NewIndex(baseDir string) (*Index, error) {
	// Create directory structure
	dirs := []string{
		filepath.Join(baseDir, "documents"),
		filepath.Join(baseDir, "indexes"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, err
		}
	}

	// Open files
	docFile, err := os.OpenFile(
		filepath.Join(baseDir, "documents", "docs.dat"),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}

	indexFile, err := os.OpenFile(
		filepath.Join(baseDir, "indexes", "inverted.idx"),
		os.O_RDWR|os.O_CREATE,
		0644,
	)
	if err != nil {
		return nil, err
	}

	idx := &Index{
		baseDir:   baseDir,
		docFile:   docFile,
		indexFile: indexFile,
		metadata: IndexMetadata{
			IndexEntries:      make(map[string][]int64),
			DocumentPositions: make(map[string]int64),
		},
	}

	// Load metadata if exists
	idx.loadMetadata()
	return idx, nil
}

func (idx *Index) loadMetadata() error {
	metaPath := filepath.Join(idx.baseDir, "metadata.json")
	data, err := os.ReadFile(metaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(data, &idx.metadata)
}

func (idx *Index) saveMetadata() error {
	metaPath := filepath.Join(idx.baseDir, "metadata.json")
	data, err := json.Marshal(idx.metadata)
	if err != nil {
		return err
	}
	return os.WriteFile(metaPath, data, 0644)
}

func (idx *Index) AddDocument(doc *Document) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Serialize and write document
	pos, err := idx.docFile.Seek(0, 2) // Seek to end
	if err != nil {
		return err
	}

	// Store document position
	idx.metadata.DocumentPositions[doc.ID] = pos

	encoder := gob.NewEncoder(idx.docFile)
	if err := encoder.Encode(doc); err != nil {
		return err
	}

	// Update inverted index
	words := strings.Fields(strings.ToLower(doc.Content))
	for _, word := range words {
		word = strings.Trim(word, ",.!? \t\n\r")
		idx.metadata.IndexEntries[word] = append(idx.metadata.IndexEntries[word], pos)
	}

	idx.metadata.DocumentCount++
	return idx.saveMetadata()
}

func (idx *Index) AddDocuments(docs []*Document) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	for _, doc := range docs {
		pos, err := idx.docFile.Seek(0, 2)
		if err != nil {
			return err
		}

		encoder := gob.NewEncoder(idx.docFile)
		if err := encoder.Encode(doc); err != nil {
			return err
		}

		idx.metadata.DocumentPositions[doc.ID] = pos
		words := strings.Fields(strings.ToLower(doc.Content))
		for _, word := range words {
			word = strings.Trim(word, ",.!? \t\n\r")
			idx.metadata.IndexEntries[word] = append(idx.metadata.IndexEntries[word], pos)
		}
		idx.metadata.DocumentCount++
	}

	return idx.saveMetadata()
}

func (idx *Index) GetDocument(id string) (*Document, error) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	pos, exists := idx.metadata.DocumentPositions[id]
	if !exists {
		return nil, nil
	}

	return idx.readDocumentAt(pos)
}

func (idx *Index) DeleteDocument(id string) error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	pos, exists := idx.metadata.DocumentPositions[id]
	if !exists {
		return nil
	}

	// Read document to get its words for index cleanup
	doc, err := idx.readDocumentAt(pos)
	if err != nil {
		return err
	}

	// Remove from inverted index
	words := strings.Fields(strings.ToLower(doc.Content))
	for _, word := range words {
		word = strings.Trim(word, ",.!? \t\n\r")
		positions := idx.metadata.IndexEntries[word]

		// Remove this document's position from the word's position list
		newPositions := make([]int64, 0, len(positions)-1)
		for _, p := range positions {
			if p != pos {
				newPositions = append(newPositions, p)
			}
		}

		if len(newPositions) == 0 {
			delete(idx.metadata.IndexEntries, word)
		} else {
			idx.metadata.IndexEntries[word] = newPositions
		}
	}

	// Remove document position
	delete(idx.metadata.DocumentPositions, id)
	idx.metadata.DocumentCount--

	return idx.saveMetadata()
}

func (idx *Index) Search(query string) ([]*Document, error) {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return nil, nil
	}

	// Get positions for all words
	wordPositions := make([]map[int64]struct{}, len(words))
	for i, word := range words {
		word = strings.Trim(word, ",.!? \t\n\r")
		positions := idx.metadata.IndexEntries[word]
		posMap := make(map[int64]struct{})
		for _, pos := range positions {
			posMap[pos] = struct{}{}
		}
		wordPositions[i] = posMap
	}

	// Find intersection of all positions
	var commonPositions []int64
	for pos := range wordPositions[0] {
		found := true
		for _, posMap := range wordPositions[1:] {
			if _, ok := posMap[pos]; !ok {
				found = false
				break
			}
		}
		if found {
			commonPositions = append(commonPositions, pos)
		}
	}

	// Read documents at common positions
	docs := make([]*Document, 0, len(commonPositions))
	for _, pos := range commonPositions {
		doc, err := idx.readDocumentAt(pos)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

func (idx *Index) readDocumentAt(pos int64) (*Document, error) {
	_, err := idx.docFile.Seek(pos, 0)
	if err != nil {
		return nil, err
	}

	decoder := gob.NewDecoder(idx.docFile)
	doc := &Document{}
	if err := decoder.Decode(doc); err != nil {
		return nil, err
	}

	return doc, nil
}

func (idx *Index) Compact() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	// Create temporary files
	tempDocPath := filepath.Join(idx.baseDir, "documents", "docs.dat.tmp")
	tempDoc, err := os.Create(tempDocPath)
	if err != nil {
		return err
	}
	defer tempDoc.Close()

	// Create new position map
	newPositions := make(map[string]int64)

	// Copy valid documents to temporary file
	for id, oldPos := range idx.metadata.DocumentPositions {
		doc, err := idx.readDocumentAt(oldPos)
		if err != nil {
			continue // Skip corrupted documents
		}

		newPos, err := tempDoc.Seek(0, 2)
		if err != nil {
			return err
		}

		encoder := gob.NewEncoder(tempDoc)
		if err := encoder.Encode(doc); err != nil {
			return err
		}

		newPositions[id] = newPos
	}

	// Update inverted index with new positions
	newIndexEntries := make(map[string][]int64)
	for word, positions := range idx.metadata.IndexEntries {
		newWordPositions := make([]int64, 0, len(positions))
		for _, oldPos := range positions {
			for id, newPos := range newPositions {
				if idx.metadata.DocumentPositions[id] == oldPos {
					newWordPositions = append(newWordPositions, newPos)
					break
				}
			}
		}
		if len(newWordPositions) > 0 {
			newIndexEntries[word] = newWordPositions
		}
	}

	// Close current file
	idx.docFile.Close()

	// Replace old file with new one
	if err := os.Rename(tempDocPath, filepath.Join(idx.baseDir, "documents", "docs.dat")); err != nil {
		return err
	}

	// Reopen file and update metadata
	idx.docFile, err = os.OpenFile(
		filepath.Join(idx.baseDir, "documents", "docs.dat"),
		os.O_RDWR,
		0644,
	)
	if err != nil {
		return err
	}

	idx.metadata.DocumentPositions = newPositions
	idx.metadata.IndexEntries = newIndexEntries

	return idx.saveMetadata()
}

func (idx *Index) Close() error {
	idx.mutex.Lock()
	defer idx.mutex.Unlock()

	if err := idx.saveMetadata(); err != nil {
		return err
	}

	if err := idx.docFile.Close(); err != nil {
		return err
	}

	return idx.indexFile.Close()
}

// Add method to get document count
func (idx *Index) DocumentCount() int {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()
	return idx.metadata.DocumentCount
}

// Add method to list all document IDs
func (idx *Index) ListDocumentIDs() []string {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	ids := make([]string, 0, len(idx.metadata.DocumentPositions))
	for id := range idx.metadata.DocumentPositions {
		ids = append(ids, id)
	}
	return ids
}

func (idx *Index) GetStats() map[string]interface{} {
	idx.mutex.RLock()
	defer idx.mutex.RUnlock()

	stats := map[string]interface{}{
		"documentCount": idx.metadata.DocumentCount,
		"uniqueWords":   len(idx.metadata.IndexEntries),
	}

	// Calculate total indexed words
	totalWords := 0
	for _, positions := range idx.metadata.IndexEntries {
		totalWords += len(positions)
	}
	stats["totalIndexedWords"] = totalWords

	return stats
}
