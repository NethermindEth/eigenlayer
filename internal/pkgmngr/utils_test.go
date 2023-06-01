package package_manager

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

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
