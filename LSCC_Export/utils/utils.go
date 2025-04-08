package utils

import (
        "os"
        "path/filepath"
)

// CreateDirIfNotExists creates a directory if it doesn't exist
func CreateDirIfNotExists(path string) error {
        if _, err := os.Stat(path); os.IsNotExist(err) {
                return os.MkdirAll(path, 0755)
        }
        return nil
}

// GetAbsolutePath returns the absolute path given a relative one
func GetAbsolutePath(relativePath string) (string, error) {
        return filepath.Abs(relativePath)
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
        info, err := os.Stat(filename)
        if os.IsNotExist(err) {
                return false
        }
        return !info.IsDir()
}