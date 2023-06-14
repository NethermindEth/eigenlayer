package package_handler

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func checkPackageFileExist(packagePath, filePath string) error {
	stats, err := os.Stat(path.Join(packagePath, filePath))
	if err != nil {
		if os.IsNotExist(err) {
			return PackageFileNotFoundError{
				fileRelativePath: filePath,
				packagePath:      packagePath,
			}
		}
		return err
	}
	if stats.IsDir() {
		return fmt.Errorf("%w: %s is not a file", ErrInvalidFilePath, filePath)
	}
	return nil
}

func checkPackageDirExist(packagePath, dirPath string) error {
	stats, err := os.Stat(path.Join(packagePath, dirPath))
	if err != nil {
		if os.IsNotExist(err) {
			return PackageDirNotFoundError{
				dirRelativePath: dirPath,
				packagePath:     packagePath,
			}
		}
		return err
	}
	if !stats.IsDir() {
		return fmt.Errorf("%w: %s is not a directory", ErrInvalidDirPath, dirPath)
	}
	return nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func packageHashes(pkgPath string) (map[string]string, error) {
	hashes := make(map[string]string, 0)

	err := filepath.Walk(filepath.Join(pkgPath, pkgDirName), func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			h, err := hashFile(path)
			if err != nil {
				return err
			}
			relativePath := strings.TrimPrefix(path, pkgPath)
			if relativePath[0] == filepath.Separator {
				relativePath = relativePath[1:]
			}
			hashes[relativePath] = h
		}
		return nil
	})
	return hashes, err
}

func parseChecksumFile(path string) (map[string]string, error) {
	checksums := make(map[string]string)

	file, err := os.Open(path)
	if err != nil {
		return checksums, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Fields(line)
		if len(parts) != 2 {
			return checksums, fmt.Errorf("invalid checksum file format")
		}
		checksums[parts[1]] = parts[0]
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return checksums, nil
}

func contains(s []string, v string) bool {
	for _, a := range s {
		if a == v {
			return true
		}
	}
	return false
}
