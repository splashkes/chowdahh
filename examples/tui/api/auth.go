package api

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	tokenDir  = ".config/chowdahh"
	tokenFile = "token"
)

// TokenPath returns the full path to the cached token file.
func TokenPath() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, tokenDir, tokenFile)
}

// LoadToken reads the cached token from disk. Returns "" if none exists.
func LoadToken() string {
	data, err := os.ReadFile(TokenPath())
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// SaveToken writes the token to disk, creating directories as needed.
func SaveToken(token string) error {
	path := TokenPath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(token+"\n"), 0600)
}

// ClearToken removes the cached token file.
func ClearToken() error {
	return os.Remove(TokenPath())
}

// ValidateTokenFormat checks the ch_person_ prefix.
func ValidateTokenFormat(token string) bool {
	return strings.HasPrefix(token, "ch_person_") || strings.HasPrefix(token, "ch_cur_")
}
