package types

import "encoding/json"

// Entry represents a stored record in a collection file.
type Entry struct {
	ID  string          `json:"_id"`
	TS  int64           `json:"ts"`
	Op  string          `json:"op"` // insert, delete, update
	Doc json.RawMessage `json:"doc"`
}

// Collection models the on-disk structure: an array of entries.
type Collection struct {
	Entries []Entry `json:"entries"`
}
