package main

import "jsondb/internal/cli"

// Database    -> is a Directory that contains `Collections` (JSON files)
// Collection  -> is a JSON file that contains multiple `Entries` (JSON objects)
// Entry       -> is a JSON object with metadata (ID, timestamp, operation) and the actual document

func main() { cli.New().Run() }
