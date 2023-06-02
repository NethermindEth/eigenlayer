package package_manager

import "errors"

var (
	ErrInvalidFilePath = errors.New("invalid file path")
	ErrInvalidDirPath  = errors.New("invalid directory path")
	ErrInvalidChecksum = errors.New("invalid checksum")
)

type PackageFileNotFoundError struct {
	fileRelativePath string
	packagePath      string
}

func (e PackageFileNotFoundError) Error() string {
	return "package file not found: " + e.fileRelativePath + " in package " + e.packagePath
}

type PackageDirNotFoundError struct {
	dirRelativePath string
	packagePath     string
}

func (e PackageDirNotFoundError) Error() string {
	return "package directory not found: " + e.dirRelativePath + " in package " + e.packagePath
}
