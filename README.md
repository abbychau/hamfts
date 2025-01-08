# Hamfts: Simple Full-Text Search Engine in Go

A lightweight implementation of an Elasticsearch-like search engine in Go, featuring inverted indexing for efficient text search capabilities.

## Features

- Document management with custom metadata
- Thread-safe operations
- Inverted index for fast full-text search
- Basic CRUD operations
- Custom document IDs and timestamps
- Metadata support for flexible document attributes

## Installation

```bash
go get github.com/abbychau/hamfts
```

## Quick Start

### Start the Server

```bash
go run main.go
```

### Basic Operations

Add a document:
```bash
curl -X POST http://localhost:8080/documents -d '{
    "id": "1",
    "content": "The quick brown fox",
    "metadata": {"category": "animals"}
}'
```

Search documents:
```bash
curl -X POST http://localhost:8080/search -d '{
    "query": "quick fox"
}'
```

Get stats:
```bash
curl http://localhost:8080/stats
```

## API Usage

### Creating and Adding Documents

```go
// Initialize a new index
idx := hamfts.NewIndex()

// Create a new document with ID and content
doc := hamfts.NewDocument("1", "The quick brown fox jumps over the lazy dog")

// Add custom metadata
doc.Metadata["category"] = "animals"

// Add document to index
idx.AddDocument(doc)
```

### Searching Documents

```go
// Simple search
results := idx.Search("quick fox")
for _, doc := range results {
    fmt.Printf("Found document: %s\n", doc.ID)
}
```

### Managing Documents

```go
// Retrieve a document by ID
doc := idx.GetDocument("1")

// Delete a document
idx.DeleteDocument("1")
```

## Document Structure

Each document contains:
- Unique ID
- Content text
- Creation timestamp
- Custom metadata map

## Thread Safety

All operations are thread-safe, protected by read-write mutex locks.

## Limitations

- Basic text search (no advanced query operations)
- In-memory storage only
- No scoring or ranking
- Simple word tokenization

## Contributing

Feel free to submit issues and enhancement requests!

## License

MIT License