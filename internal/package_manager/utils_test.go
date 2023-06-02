package package_manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/NethermindEth/eigen-wiz/internal/package_manager/testdata"
	"github.com/stretchr/testify/assert"
)

func FuzzHashFile(f *testing.F) {
	for i := 0; i < 10; i++ {
		f.Add([]byte(fmt.Sprintf("file content %d\n", i)))
	}

	filePath := filepath.Join(f.TempDir(), "file.txt")
	file, err := os.Create(filePath)
	if err != nil {
		f.Fatalf("failed to create temp file: %v", err)
	}
	defer file.Close()

	f.Fuzz(func(t *testing.T, fileContent []byte) {
		if _, err := file.Write(fileContent); err != nil {
			t.Fatalf("failed to write to temp file: %v", err)
		}
		sha256sum := exec.Command("sha256sum", filePath)
		output, err := sha256sum.Output()
		if err != nil {
			t.Fatalf("failed to run sha256sum: %v", err)
		}
		fileHash, err := hashFile(filePath)
		assert.NoError(t, err)
		assert.Equal(t, strings.Split(string(output), " ")[0], fileHash)
	})
}

func TestCheckPackageFileExist(t *testing.T) {
	testDir := t.TempDir()
	testdata.SetupDir(t, "mock-avs", testDir)

	ts := []struct {
		name     string
		filePath string
		err      error
	}{
		{
			name:     "file exists",
			filePath: "pkg/manifest.yml",
			err:      nil,
		},
		{
			name:     "file does not exist",
			filePath: "pkg/manifest2.yml",
			err: PackageFileNotFoundError{
				fileRelativePath: "pkg/manifest2.yml",
				packagePath:      filepath.Join(testDir, "mock-avs"),
			},
		},
		{
			name:     "is not a file",
			filePath: "pkg",
			err:      ErrInvalidFilePath,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPackageFileExist(filepath.Join(testDir, "mock-avs"), tc.filePath)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}

func TestCheckPackageDirExist(t *testing.T) {
	testDir := t.TempDir()
	testdata.SetupDir(t, "mock-avs", testDir)

	ts := []struct {
		name    string
		dirPath string
		err     error
	}{
		{
			name:    "dir exists",
			dirPath: "pkg",
			err:     nil,
		},
		{
			name:    "does not exist",
			dirPath: "pkg2",
			err: PackageDirNotFoundError{
				dirRelativePath: "pkg2",
				packagePath:     filepath.Join(testDir, "mock-avs"),
			},
		},
		{
			name:    "is not a directory",
			dirPath: "pkg/manifest.yml",
			err:     ErrInvalidDirPath,
		},
	}
	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			err := checkPackageDirExist(filepath.Join(testDir, "mock-avs"), tc.dirPath)
			assert.ErrorIs(t, err, tc.err)
		})
	}
}
