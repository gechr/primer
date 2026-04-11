package human

import (
	"os"
	"path/filepath"
	"strings"
)

// ContractHome replaces the user's home directory prefix with ~.
func ContractHome(path string) string {
	if home, err := os.UserHomeDir(); err == nil {
		if path == home {
			return "~"
		}
		if rest, ok := strings.CutPrefix(path, home+"/"); ok {
			return "~/" + rest
		}
	}
	return path
}

// ExpandPath expands ~ to the user's home directory and resolves
// environment variables via os.ExpandEnv.
func ExpandPath(path string) string {
	if path == "" {
		return path
	}
	if path == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
	}
	if rest, ok := strings.CutPrefix(path, "~/"); ok {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, rest)
		}
	}
	return os.ExpandEnv(path)
}
