package utils

import (
	"archive/tar"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompressToTarGz(t *testing.T) {
	testDir := t.TempDir()
	pkgDir := filepath.Join(testDir, "mock-avs")
	outTarPath := filepath.Join(testDir, "out.tar.gz")
	outTarContentDir := filepath.Join(testDir, "out")

	err := os.MkdirAll(pkgDir, 0o755)
	require.NoError(t, err, "failed to create mock-avs dir")
	err = exec.Command("git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir).Run()
	require.NoError(t, err, "failed to clone mock-avs repo")

	outTar, err := os.OpenFile(outTarPath, os.O_CREATE|os.O_RDWR, 0o755)
	require.NoError(t, err)

	err = CompressToTarGz(pkgDir, outTar)
	require.NoError(t, err)

	err = os.MkdirAll(outTarContentDir, 0o755)
	require.NoError(t, err, "failed to create out dir")
	err = exec.Command("tar", "-xf", outTarPath, "-C", outTarContentDir).Run()
	require.NoError(t, err, "failed to create mock-avs.tar.gz")

	assertEqualDirs(t, pkgDir, outTarContentDir)
}

func TestDecompressTarGz(t *testing.T) {
	testDir := t.TempDir()
	pkgDir := filepath.Join(testDir, "mock-avs")
	tarPath := filepath.Join(testDir, "mock-avs.tar.gz")
	outDir := filepath.Join(testDir, "out")

	err := os.MkdirAll(pkgDir, 0o755)
	require.NoError(t, err, "failed to create mock-avs dir")
	err = exec.Command("git", "clone", "--single-branch", "-b", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo(), pkgDir).Run()
	require.NoError(t, err, "failed to clone mock-avs repo")

	err = exec.Command("tar", "-czf", tarPath, "-C", pkgDir, ".").Run()
	require.NoError(t, err, "failed to create mock-avs.tar.gz")

	tarFile, err := os.Open(tarPath)
	require.NoError(t, err, "failed to open mock-avs.tar.gz")

	err = DecompressTarGz(tarFile, outDir)
	require.NoError(t, err, "failed to decompress mock-avs.tar.gz")

	assertEqualDirs(t, pkgDir, outDir)
}

func TestTarInit(t *testing.T) {
	fs := afero.NewMemMapFs()
	tarPath := "/init.tar"

	err := TarInit(fs, tarPath)
	require.NoError(t, err, "failed to init tar file")

	tarStat, err := fs.Stat(tarPath)
	require.NoError(t, err, "failed to stat tar file")

	assert.Equal(t, os.FileMode(0o644), tarStat.Mode(), "tar file has wrong mode")
	assert.Equal(t, int64(2*tarBlockSize), tarStat.Size(), "tar file has wrong size")
}

func TestTarPrepareToAppend(t *testing.T) {
	tc := []struct {
		name    string
		tarFile func(*testing.T) afero.File
		err     error
	}{
		{
			name: "empty tar file",
			tarFile: func(t *testing.T) afero.File {
				fs := afero.NewMemMapFs()
				tarPath := "/empty.tar"

				tarFile, err := fs.Create(tarPath)
				require.NoError(t, err, "failed to create tar file")

				d := make([]byte, 1024)
				_, err = tarFile.Write(d)
				require.NoError(t, err, "failed to write to tar file")
				err = tarFile.Close()
				require.NoError(t, err, "failed to close tar file")

				tarFile, err = fs.Open(tarPath)
				require.NoError(t, err, "failed to open tar file")
				return tarFile
			},
		},
		{
			name: "error, tar file with less than 2 blocks (1024 bytes)",
			tarFile: func(t *testing.T) afero.File {
				fs := afero.NewMemMapFs()
				tarPath := "/empty.tar"

				tarFile, err := fs.Create(tarPath)
				require.NoError(t, err, "failed to create tar file")

				d := make([]byte, 1000)
				_, err = tarFile.Write(d)
				require.NoError(t, err, "failed to write to tar file")
				err = tarFile.Close()
				require.NoError(t, err, "failed to close tar file")

				tarFile, err = fs.Open(tarPath)
				require.NoError(t, err, "failed to open tar file")
				return tarFile
			},
			err: fmt.Errorf("%w: tar file is not empty but has less than 2 blocks (1024 bytes)", ErrTarPrepareToAppend),
		},
		{
			name: "error, tar file with 2 blocks (1024 bytes) but not zeroed",
			tarFile: func(t *testing.T) afero.File {
				fs := afero.NewMemMapFs()
				tarPath := "/empty.tar"

				tarFile, err := fs.Create(tarPath)
				require.NoError(t, err, "failed to create tar file")

				d := make([]byte, 1024)
				d[100] = 1
				_, err = tarFile.Write(d)
				require.NoError(t, err, "failed to write to tar file")
				err = tarFile.Close()
				require.NoError(t, err, "failed to close tar file")

				tarFile, err = fs.Open(tarPath)
				require.NoError(t, err, "failed to open tar file")
				return tarFile
			},
			err: fmt.Errorf("%w: last 1024 bytes are not all 0", ErrTarPrepareToAppend),
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			tarFile := tt.tarFile(t)
			defer tarFile.Close()

			err := TarPrepareToAppend(tarFile)
			if tt.err == nil {
				assert.NoError(t, err, "unexpected error")
			} else {
				require.Error(t, err, "expected error")
				assert.EqualError(t, err, tt.err.Error(), "not equal errors")
			}
		})
	}
}

func TestTarAddDir(t *testing.T) {
	testDir := t.TempDir()
	srcPath := filepath.Join(testDir, "src")
	err := os.MkdirAll(srcPath, 0o755)
	require.NoError(t, err, "failed to create src dir")

	// create a file in the src directory
	fileContent := []byte("test file content")
	filePath := filepath.Join(srcPath, "test.txt")
	file, err := os.Create(filePath)
	require.NoError(t, err, "failed to create test file")
	n, err := file.Write(fileContent)
	require.NoError(t, err, "failed to write to test file")
	require.Equal(t, len(fileContent), n, "failed to write all data to test file")

	// create a tar file and add the src directory to it
	tarPath := filepath.Join(testDir, "test.tar")
	tarFile, err := os.Create(tarPath)
	require.NoError(t, err, "failed to create tar file")
	defer tarFile.Close()

	err = TarAddDir(srcPath, "prefix/path", tarFile)
	require.NoError(t, err, "failed to add src dir to tar file")

	// read the tar file and check its contents
	tarFile, err = os.Open(tarPath)
	require.NoError(t, err, "failed to open tar file")
	defer tarFile.Close()

	tarReader := tar.NewReader(tarFile)
	// Root header
	header, err := tarReader.Next()
	require.NoError(t, err, "failed to read tar header")
	assert.Equal(t, "prefix/path", header.Name, "root header has wrong name")
	// Test file header
	header, err = tarReader.Next()
	require.NoError(t, err, "failed to read tar header")
	assert.Equal(t, "prefix/path/test.txt", header.Name, "test.txt header has wrong name")
	assert.Equal(t, int64(len(fileContent)), header.Size, "test.txt header has wrong size")
	// Test file content
	content := make([]byte, len(fileContent))
	n, err = tarReader.Read(content)
	assert.Equal(t, n, len(fileContent), "failed to read all data from test.txt")
	require.ErrorIs(t, err, io.EOF, "expected EOF")
	// Test EOF
	_, err = tarReader.Next()
	require.Equal(t, io.EOF, err, "expected EOF")
}

func TestTarAddFile(t *testing.T) {
	// Create temporary directory and file
	tmpDir := t.TempDir()
	tmpFile, err := os.CreateTemp(tmpDir, "test-*.txt")
	require.NoError(t, err)
	_, err = tmpFile.WriteString("This is a test file")
	require.NoError(t, err)

	tests := []struct {
		name     string
		src      string
		dest     string
		expected []byte
		wantErr  bool
	}{
		{
			name:     "add file to tar",
			src:      tmpFile.Name(),
			dest:     "testfile.txt",
			expected: []byte("This is a test file"),
			wantErr:  false,
		},
		{
			name:     "add nonexistent file to tar",
			src:      filepath.Join(tmpDir, "nonexistent.txt"),
			dest:     "nonexistent.txt",
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "add directory to tar",
			src:      tmpDir,
			dest:     "testdata",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary tar file
			tmpTarFile, err := os.CreateTemp(t.TempDir(), "test-*.tar")
			require.NoError(t, err)
			defer os.Remove(tmpTarFile.Name())

			// Add file to tar
			err = TarAddFile(tt.src, tt.dest, tmpTarFile)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			// Read tar file
			tmpTarFile.Seek(0, 0)
			tr := tar.NewReader(tmpTarFile)

			// Check tar contents
			header, err := tr.Next()
			require.NoError(t, err)
			assert.Equal(t, tt.dest, header.Name)

			buf := new(bytes.Buffer)
			_, err = io.Copy(buf, tr)
			require.NoError(t, err)
			assert.Equal(t, tt.expected, buf.Bytes())
		})
	}
}

func assertEqualDirs(t *testing.T, dir1, dir2 string) {
	err := filepath.Walk(dir1, func(path1 string, info1 os.FileInfo, err1 error) error {
		if err1 != nil {
			return err1
		}

		path2 := filepath.Join(dir2, path1[len(dir1):])

		if info1.IsDir() {
			assert.DirExists(t, path2)
		} else {
			assert.FileExists(t, path2)
			assertEqualFiles(t, path1, path2)
		}
		return nil
	})
	require.NoError(t, err, "failed to walk dir %s", dir1)
}

func assertEqualFiles(t *testing.T, f1, f2 string) {
	file1, err := os.ReadFile(f1)
	require.NoError(t, err, "failed to read file %s", f1)
	file2, err := os.ReadFile(f2)
	require.NoError(t, err, "failed to read file %s", f2)
	assert.Equal(t, file1, file2)
}
