package helper

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type FileType string

const (
	TypeFile      FileType = "file"
	TypeDirectory FileType = "dir"
)

// PathExist checks if a given path exists and whether it's a file or directory
func PathExist(path string) (bool, FileType, error) {
	info, err := os.Stat(path)
	if err == nil {
		// Path exists
		if info.IsDir() {
			return true, TypeDirectory, nil
		}
		return true, TypeFile, nil
	}

	// Does not exist
	if os.IsNotExist(err) {
		return false, "", nil
	}

	// other errors: permission denied, etc
	return false, "", err
}

// ValidatePath checks for directory traversal and invalid names
func ValidatePath(name string) error {
	if name == "" {
		return fmt.Errorf("name is empty")
	}
	// disallow path separators and '..'
	if strings.Contains(name, string(filepath.Separator)) || strings.Contains(name, "..") {
		return fmt.Errorf("invalid characters in name")
	}
	// optional: allow only alphanumeric, underscore, dash
	ok, _ := regexp.MatchString(`^[A-Za-z0-9_-]+$`, name)
	if !ok {
		return fmt.Errorf("name must match [A-Za-z0-9_-]+")
	}
	return nil
}
