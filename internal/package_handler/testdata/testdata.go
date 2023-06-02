package testdata

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
)

//go:embed *
var TestData embed.FS

// SetupDir copies the testdata directory to the dest directory. testDataPath
// is a relative path to the testdata directory.
func SetupDir(t *testing.T, testDataPath string, dest string) {
	t.Helper()
	err := fs.WalkDir(TestData, testDataPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if err := os.MkdirAll(filepath.Join(dest, path), 0755); err != nil {
				return err
			}
		} else {
			// If the entry is a file, copy it to the temp dir
			data, err := fs.ReadFile(TestData, path)
			if err != nil {
				return err
			}
			destFile, err := os.Create(filepath.Join(dest, path))
			if err != nil {
				return err
			}
			if _, err := destFile.Write(data); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to setup test data directory: %v", err)
	}
}
