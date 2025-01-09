package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"hamfts/client"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Server URL")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: hamctl <command> [args...]")
		fmt.Println("Commands:")
		fmt.Println("  search <query>")
		fmt.Println("  add <id> <content> [metadata]")
		fmt.Println("  list")
		fmt.Println("  delete <id>")
		fmt.Println("  stats")
		os.Exit(1)
	}

	c := client.NewClient(*serverURL)
	cmd := flag.Args()[0]

	switch cmd {
	case "search":
		if len(flag.Args()) < 2 {
			fmt.Println("Usage: hamctl search <query>")
			os.Exit(1)
		}
		results, err := c.Search(flag.Args()[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Search failed: %v\n", err)
			os.Exit(1)
		}
		printJSON(results)

	case "add":
		if len(flag.Args()) < 3 {
			fmt.Println("Usage: hamctl add <id> <content> [metadata]")
			os.Exit(1)
		}

		var metadata map[string]interface{}
		if len(flag.Args()) > 3 {
			if err := json.Unmarshal([]byte(flag.Args()[3]), &metadata); err != nil {
				fmt.Fprintf(os.Stderr, "Invalid metadata JSON: %v\n", err)
				os.Exit(1)
			}
		}

		err := c.AddDocument(flag.Args()[1], flag.Args()[2], metadata)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Add document failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Document added successfully")

	case "list":
		docs, err := c.ListDocuments()
		if err != nil {
			fmt.Fprintf(os.Stderr, "List documents failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(strings.Join(docs, "\n"))

	case "delete":
		if len(flag.Args()) < 2 {
			fmt.Println("Usage: hamctl delete <id>")
			os.Exit(1)
		}
		err := c.DeleteDocument(flag.Args()[1])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Delete document failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Document deleted successfully")

	case "stats":
		stats, err := c.GetStats()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Get stats failed: %v\n", err)
			os.Exit(1)
		}
		printJSON(stats)

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}

func printJSON(v interface{}) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	encoder.Encode(v)
}
