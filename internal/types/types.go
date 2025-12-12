package types

import "encoding/json"

type Operation string

const (
	OpInsert Operation = "insert"
	OpQuery  Operation = "query"
	OpDelete Operation = "delete"
	OpUpdate Operation = "update"
)

// Record is parsed from user input into an Entry
type Record struct {
	Data map[string]any `json:"data"`
}

// Entry represents a stored record in a collection file.
type Entry struct {
	ID  string          `json:"_id"` // Primary key
	TS  int64           `json:"ts"`  // Unix timestamp in nanoseconds
	Op  Operation       `json:"op"`  // insert, delete, update
	Doc json.RawMessage `json:"doc"` // doc content in JSON format. Nil for delete operations.
}

// Collection models the on-disk structure: an array of entries.
type Collection struct {
	Entries []Entry `json:"entries"`
}
