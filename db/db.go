package db

import (
	"encoding/json"
	"fmt"
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

func (db *DB) CreateCollection(name string) error {
	path := db.collectionPath(name)

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

	// Collection does not exist. Proceed
	col, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o666)
	if err != nil {
		return fmt.Errorf("failed to create collection file %q: %w", path, err)
	}
	defer col.Close()

	// Initialize with an empty CollectionFile JSON
	if err := json.NewEncoder(col).Encode(types.CollectionFile{Entries: nil}); err != nil {
		return fmt.Errorf("failed to write initial collection content to %q: %w", path, err)
	}

	return nil
}

// ReadCollection (Read All) returns all entries from a collection file
func (db *DB) ReadCollection(name string) (*types.CollectionFile, error) {
	path := db.collectionPath(name)

	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &types.CollectionFile{}, fmt.Errorf("collection %q does not exist", name)
		}
		return &types.CollectionFile{}, err
	}
	defer file.Close()

	collections := types.CollectionFile{}
	if err = json.NewDecoder(file).Decode(&collections); err != nil {
		return &types.CollectionFile{}, err
	}

	return &collections, nil
}
