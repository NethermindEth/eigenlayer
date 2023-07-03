package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func dataDirPath(t *testing.T) string {
	t.Helper()
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		if err != nil {
			t.Fatal(err)
		}
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	dataDir := filepath.Join(userDataHome, ".eigen")
	return dataDir
}
