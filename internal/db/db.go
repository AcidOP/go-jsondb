package db

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	helper "jsondb/internal/helpers"
	"jsondb/internal/types"
	"os"
	"path/filepath"
	"time"
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
	initial := types.Collection{Entries: []types.Entry{}}
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
func (db *DB) ReadCollection(name string) (*types.Collection, error) {
	path := db.collectionPath(name)

	if err := helper.ValidatePath(name); err != nil {
		return &types.Collection{}, fmt.Errorf("invalid collection name: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.Collection{}, fmt.Errorf("collection %q does not exist", name)
		}
		return &types.Collection{}, err
	}
	defer file.Close()

	collections := types.Collection{}
	decoder := json.NewDecoder(file)

	if err = decoder.Decode(&collections); err != nil {
		if err == io.EOF {
			// empty file -> empty collection
			return &types.Collection{Entries: []types.Entry{}}, nil
		}
		return nil, fmt.Errorf("decode collection JSON: %w", err)
	}

	if collections.Entries == nil {
		collections.Entries = []types.Entry{}
	}

	return &collections, nil
}

func (db *DB) InsertRecord(collectionName string, data *types.Record) error {
	if data == nil || data.Data == nil {
		return fmt.Errorf("data cannot be nil")
	}
	if collectionName == "" {
		return fmt.Errorf("collection name cannot be empty")
	}

	path := db.collectionPath(collectionName)

	if err := helper.ValidatePath(collectionName); err != nil {
		return fmt.Errorf("invalid collection name: %v", err)
	}

	collection, err := db.ReadCollection(collectionName)
	if err != nil {
		return fmt.Errorf("failed to read collection %q: %w", collectionName, err)
	}

	entry, err := db.recordToEntry(types.OpInsert, data)
	if err != nil {
		return fmt.Errorf("failed to convert record to entry: %w", err)
	}

	collection.Entries = append(collection.Entries, *entry)

	// Backup the collection table so we can revert in case of write failure
	// Atomicity on the way in :)
	backupPath := path + ".tmp"

	file, err := os.Create(backupPath)
	if err != nil {
		file.Close()
		return err
	}

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(collection); err != nil {
		file.Close()
		os.Remove(backupPath)
		return fmt.Errorf("failed to write collection JSON: %w", err)
	}

	// Write to disk
	if err := file.Sync(); err != nil {
		file.Close()
		os.Remove(backupPath)
		return fmt.Errorf("failed to sync temp file: %w", err)
	}

	// Close file so windows/unix will allow rename
	if err := file.Close(); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to close temp file: %w", err)
	}

	// Move backup to original path
	if err := os.Rename(backupPath, path); err != nil {
		os.Remove(backupPath)
		return fmt.Errorf("failed to write updated collection: %w", err)
	}

	return nil
}

// recordToEntry generates ID, timestamp, and marshalls the data and returns an DB ready Entry
func (db *DB) recordToEntry(op types.Operation, data *types.Record) (*types.Entry, error) {
	if data == nil || data.Data == nil {
		return nil, fmt.Errorf("data cannot be nil")
	}

	idBytes := make([]byte, 16)
	if _, err := rand.Read(idBytes); err != nil {
		return nil, fmt.Errorf("failed to generate id: %w", err)
	}

	doc, err := json.Marshal(data.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal data to JSON: %w", err)
	}

	e := &types.Entry{
		ID:  hex.EncodeToString(idBytes),
		TS:  time.Now().UnixNano(),
		Op:  op,
		Doc: doc,
	}

	return e, nil
}
