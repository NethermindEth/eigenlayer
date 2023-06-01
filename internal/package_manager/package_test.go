package package_manager

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck(t *testing.T) {
	type testCase struct {
		name      string
		pkgFolder string
		err       error
	}
	ts := []testCase{
		func() testCase {
			return testCase{
				name:      "valid package",
				pkgFolder: setupPackage(t),
				err:       nil,
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "pkg")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "pkg folder does not exist",
				pkgFolder: pkgFolder,
				err: PackageDirNotFoundError{
					dirRelativePath: "pkg",
					packagePath:     pkgFolder,
				},
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "checksum.txt")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "checksum.txt file does not exist",
				pkgFolder: pkgFolder,
				err: PackageFileNotFoundError{
					fileRelativePath: "checksum.txt",
					packagePath:      pkgFolder,
				},
			}
		}(),
		func() testCase {
			pkgFolder := setupPackage(t)
			if err := exec.Command("rm", "-rf", filepath.Join(pkgFolder, "pkg", "manifest.yml")).Run(); err != nil {
				t.Fatal("error preparing the test: " + err.Error())
			}
			return testCase{
				name:      "missing file that is listed in checksum.txt produces ErrInvalidChecksum",
				pkgFolder: pkgFolder,
				err:       fmt.Errorf("%w: expected 4 files, got 3", ErrInvalidChecksum),
			}
		}(),
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			pkgManager := NewPackageManager(tc.pkgFolder)
			err := pkgManager.Check()
			assert.Equal(t, tc.err, err)
		})
	}
}

func setupPackage(t *testing.T) string {
	t.Helper()
	pkgFolder := t.TempDir()

	mockTapRepo := "https://github.com/NethermindEth/mock-avs.git"
	tag := "v0.1.0"

	t.Logf("Cloning mock tap repo %s and tag %s into %s", mockTapRepo, tag, pkgFolder)

	if err := exec.Command("git", "clone", "--single-branch", "-b", tag, mockTapRepo, pkgFolder).Run(); err != nil {
		t.Fatal("error cloning the mock tap repo: " + err.Error())
	}
	return pkgFolder
}
