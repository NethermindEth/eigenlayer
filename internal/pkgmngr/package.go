package package_manager

import (
	"fmt"
	"path/filepath"
)

const (
	pkgDirName       = "pkg"
	checksumFileName = "checksum.txt"
)

type PackageManager struct {
	path string
}

func NewPackageManager(path string) *PackageManager {
	return &PackageManager{path: path}
}

func (p *PackageManager) Check() error {
	if err := checkPackageDirExist(p.path, pkgDirName); err != nil {
		return err
	}
	if err := checkPackageFileExist(p.path, checksumFileName); err != nil {
		return err
	}

	return p.checkSum()
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
