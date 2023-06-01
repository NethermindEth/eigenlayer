package package_manager

import (
	"errors"
	"fmt"
	"path/filepath"
)

const (
	pkgDirName       = "pkg"
	checksumFileName = "checksum.txt"
)

// PackageManager is used to interact with an AVS node software package at the given
// path.
type PackageManager struct {
	path string
}

// NewPackageManager creates a new PackageManager instance for the given package path.
func NewPackageManager(path string) *PackageManager {
	return &PackageManager{path: path}
}

// Check validates a package. It returns an error if the package is invalid.
// It checks the existence of some required files and directories and computes the
// checksums comparing them with the ones listed in the checksum.txt file.
func (p *PackageManager) Check() error {
	if err := checkPackageDirExist(p.path, pkgDirName); err != nil {
		return err
	}
	err := checkPackageFileExist(p.path, checksumFileName)
	if err != nil {
		var fileNotFoundErr PackageFileNotFoundError
		if errors.As(err, &fileNotFoundErr) {
			return nil
		}
		return err
	} else {
		return p.checkSum()
	}
}

func (p *PackageManager) checkSum() error {
	currentChecksums, err := parseChecksumFile(filepath.Join(p.path, checksumFileName))
	if err != nil {
		return err
	}
	computedChecksums, err := packageHashes(p.path)
	if err != nil {
		return err
	}
	if len(currentChecksums) != len(computedChecksums) {
		return fmt.Errorf("%w: expected %d files, got %d", ErrInvalidChecksum, len(currentChecksums), len(computedChecksums))
	}
	for file, hash := range currentChecksums {
		if computedChecksums[file] != hash {
			return fmt.Errorf("%w: checksum mismatch for file %s, expected %s, got %s", ErrInvalidChecksum, file, hash, computedChecksums[file])
		}
	}
	return nil
}
