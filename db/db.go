package db

import (
	"encoding/json"
	"fmt"
	"io"
	helper "jsondb/helpers"
	"jsondb/types"
	"os"
	"path/filepath"
)

type DB struct {
	BaseDir string
}

func New(baseDir string) *DB { return &DB{BaseDir: baseDir} }

// Init creates the dir (database) provided
func (db *DB) Init() error {
	exists, _, err := helper.PathExist(db.BaseDir)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("database with name %q already exists. \nUse \"load\" command to open it", db.BaseDir)
	}

	return os.MkdirAll(db.BaseDir, 0o755)
}

// CollectionPath returns the full path to a collection file
func (db *DB) collectionPath(name string) string { return filepath.Join(db.BaseDir, (name + ".json")) }

// CreateCollection creates a new collection (JSON file) in the database directory
// Returns error if collection already exists
func (db *DB) CreateCollection(name string) error {
	path := db.collectionPath(name)

	if err := helper.ValidatePath(name); err != nil {
		return fmt.Errorf("invalid collection name: %v", err)
	}

	exists, fileType, err := helper.PathExist(db.collectionPath(name))
	if err != nil {
		return err
	}
	if exists {
		if fileType == "dir" {
			return fmt.Errorf("a directory exists at %q; cannot create collection with that name", path)
		}
		return fmt.Errorf("collection %q already exists", name)
	}

	// Create parent dir (in case BaseDir was created but missing permissions)
	if err := os.MkdirAll(db.BaseDir, 0o755); err != nil {
		return fmt.Errorf("failed to create database directory: %w", err)
	}

	// Collection does not exist.
	// Write an empty collection structure
	initial := types.CollectionFile{Entries: []types.Entry{}}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create collection file %q: %w", path, err)
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(initial); err != nil {
		f.Close()
		return fmt.Errorf("failed to write initial collection: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("failed to close collection file: %w", err)
	}

	return nil
}

// ReadCollection (Read All) returns all entries from a collection file
// Empty collection will return an empty CollectionFile struct
// Error is returned if collection does not exist or on read/parse failure
func (db *DB) ReadCollection(name string) (*types.CollectionFile, error) {
	path := db.collectionPath(name)

	if err := helper.ValidatePath(name); err != nil {
		return &types.CollectionFile{}, fmt.Errorf("invalid collection name: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.CollectionFile{}, fmt.Errorf("collection %q does not exist", name)
		}
		return &types.CollectionFile{}, err
	}
	defer file.Close()

	collections := types.CollectionFile{}
	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&collections); err != nil {
		if err == io.EOF {
			// empty file -> empty collection
			return &types.CollectionFile{Entries: []types.Entry{}}, nil
		}
		return nil, fmt.Errorf("decode collection JSON: %w", err)
	}

	if collections.Entries == nil {
		collections.Entries = []types.Entry{}
	}

	return &collections, nil
}
