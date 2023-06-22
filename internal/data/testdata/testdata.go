package testdata

import (
	"embed"
	"io/fs"
	"testing"
)

//go:embed *
var TestData embed.FS

func SetupProfileFS(t *testing.T, instanceName string) fs.FS {
	t.Helper()
	instanceFs, err := fs.Sub(TestData, instanceName)
	if err != nil {
		t.Fatalf("failed to setup instance filesystem: %v", err)
	}
	return instanceFs
}
