package package_handler

import (
	"errors"
	"strings"
)

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

type InvalidManifestError struct {
	message       string
	invalidFields []string
	missingFields []string
}

func (e InvalidManifestError) Error() string {
	// Nil error
	if e.message == "" {
		return ""
	}

	msg := e.message + " -> "
	if len(e.invalidFields) > 0 {
		msg += "invalid fields: " + strings.Join(e.invalidFields, ", ") + ". "
	}
	if len(e.missingFields) > 0 {
		msg += "missing fields: " + strings.Join(e.missingFields, ", ") + ". "
	}
	return msg
}
