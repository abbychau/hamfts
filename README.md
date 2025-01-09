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

// ...existing code...

### Searching Documents