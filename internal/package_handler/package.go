package package_handler

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	pkgDirName       = "pkg"
	checksumFileName = "checksum.txt"
	manifestFileName = "manifest.yml"
	profileFileName  = "profile.yml"
)

// PackageHandler is used to interact with an AVS node software package at the given
// path.
type PackageHandler struct {
	path string
}

// NewPackageHandler creates a new PackageHandler instance for the given package path.
func NewPackageHandler(path string) *PackageHandler {
	return &PackageHandler{path: path}
}

// Check validates a package. It returns an error if the package is invalid.
// It checks the existence of some required files and directories and computes the
// checksums comparing them with the ones listed in the checksum.txt file.
func (p *PackageHandler) Check() error {
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

func (p *PackageHandler) checkSum() error {
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

// Profiles returns the list of profiles defined in the package.
func (p *PackageHandler) Profiles() ([]Profile, error) {
	names, err := p.profilesNames()
	if err != nil {
		return nil, err
	}

	profiles := make([]Profile, 0)
	for _, profileName := range names {
		profile, err := p.parseProfile(profileName)
		if err != nil {
			return nil, err
		}

		if err := profile.Validate(); err != nil {
			return nil, err
		}

		profiles = append(profiles, *profile)
	}

	return profiles, nil
}

func (p *PackageHandler) parseManifest() (*Manifest, error) {
	manifestPath := filepath.Join(p.path, pkgDirName, manifestFileName)
	// Read the manifest file
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ReadingManifestError{
			pkgPath: p.path,
		}, err)
	}

	var manifest Manifest
	err = yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ParsingManifestError{
			pkgPath: p.path,
		}, err)
	}

	return &manifest, nil
}

func (p *PackageHandler) profilesNames() ([]string, error) {
	manifest, err := p.parseManifest()
	if err != nil {
		return nil, err
	}

	if err := manifest.validate(); err != nil {
		return nil, err
	}

	names := make([]string, len(manifest.Profiles))
	for i, profile := range manifest.Profiles {
		names[i] = profile.Name
	}

	return names, nil
}

func (p *PackageHandler) parseProfile(profileName string) (*Profile, error) {
	data, err := os.ReadFile(filepath.Join(p.path, pkgDirName, profileName, profileFileName))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ReadingProfileError{
			profileName: profileName,
		}, err)
	}

	var profile Profile
	err = yaml.Unmarshal(data, &profile)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ParsingProfileError{
			profileName: profileName,
		}, err)
	}

	return &profile, nil
}

// DotEnv returns the .env file for the given profile.
// Assumes the package has been checked and is valid.
func (p *PackageHandler) DotEnv(profile string) (map[string]string, error) {
	env := make(map[string]string)
	envPath := filepath.Join(p.path, pkgDirName, profile, ".env")

	data, err := os.ReadFile(envPath)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ReadingDotEnvError{
			pkgPath:     p.path,
			profileName: profile,
		}, err)
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "=")
		if len(parts) != 2 {
			continue
		}
		env[parts[0]] = parts[1]
	}
	return env, nil
}
