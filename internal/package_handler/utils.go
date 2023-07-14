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

	"github.com/go-git/go-git/v5"
	"github.com/spf13/afero"
)

func checkPackageFileExist(packagePath, filePath string, afs afero.Fs) error {
	stats, err := afs.Stat(path.Join(packagePath, filePath))
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

func checkPackageDirExist(packagePath, dirPath string, afs afero.Fs) error {
	stats, err := afs.Stat(path.Join(packagePath, dirPath))
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

func hashFile(path string, afs afero.Fs) (hash string, err error) {
	file, err := afs.Open(path)
	if err != nil {
		return "", err
	}
	defer func() {
		cerr := file.Close()
		if err == nil {
			err = cerr
		}
	}()

	h := sha256.New()
	if _, err := io.Copy(h, file); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func packageHashes(pkgPath string, afs afero.Fs) (map[string]string, error) {
	hashes := make(map[string]string, 0)

	err := afero.Walk(afs, filepath.Join(pkgPath, pkgDirName), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			h, err := hashFile(path, afs)
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

func parseChecksumFile(path string, afs afero.Fs) (map[string]string, error) {
	checksums := make(map[string]string)

	file, err := afs.Open(path)
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

func cloneRepository(dst string, afs afero.Fs, o *git.CloneOptions) (*git.Repository, error) {
	// Create temp directory
	tempDir, err := os.MkdirTemp("", "egn-temp-repo")
	if err != nil {
		return nil, err
	}

	repo, err := git.PlainClone(tempDir, false, o)
	if err != nil {
		return nil, err
	}

	// Copy files from temp directory to path directory
	return repo, filepath.Walk(tempDir, func(path string, info os.FileInfo, ierr error) (err error) {
		if ierr != nil {
			return err
		}
		relativePath := strings.TrimPrefix(path, tempDir)
		if len(relativePath) > 1 && relativePath[0] == filepath.Separator {
			relativePath = relativePath[1:]
		}
		if info.IsDir() {
			// Ignore .git directory
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			err = afs.MkdirAll(filepath.Join(dst, relativePath), os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer func() {
				cerr := file.Close()
				if err == nil {
					err = cerr
				}
			}()
			newFile, err := afs.Create(filepath.Join(dst, relativePath))
			if err != nil {
				return err
			}
			defer func() {
				cerr := newFile.Close()
				if err == nil {
					err = cerr
				}
			}()
			_, err = io.Copy(newFile, file)
			if err != nil {
				return err
			}
		}
		return nil
	})
}
