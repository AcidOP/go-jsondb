package helper

import (
	"os"
)

// PathExist checks if a given path exists and whether it's a file or directory
func PathExist(path string) (bool, string, error) {
	info, err := os.Stat(path)
	if err == nil {
		// Path exists
		if info.IsDir() {
			return true, "dir", nil
		}
		return true, "file", nil
	}

	// Does not exist
	if os.IsNotExist(err) {
		return false, "", nil
	}

	// other errors: permission denied, etc
	return false, "", err
}
