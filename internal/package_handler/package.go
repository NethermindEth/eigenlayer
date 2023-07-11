package package_handler

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

const (
	pkgDirName       = "pkg"
	checksumFileName = "checksum.txt"
	manifestFileName = "manifest.yml"
	profileFileName  = "profile.yml"
)

var tagVersionRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// PackageHandler is used to interact with an AVS node software package at the given
// path.
type PackageHandler struct {
	path string
}

// NewPackageHandler creates a new PackageHandler instance for the given package path.
func NewPackageHandler(path string) *PackageHandler {
	return &PackageHandler{path: path}
}

// NewPackageHandlerOptions is used to provide options to the NewPackageHandlerFromURL
type NewPackageHandlerOptions struct {
	// Path is the path where the package will be cloned
	Path string
	// URL is the URL of the git repository
	URL string
	// GitAuth is used to provide authentication to a private git repository
	GitAuth *GitAuth
}

// GitAuth is used to provide authentication to a private git repository. Two types of
// authentication are supported (tested with GitHub):
//
//  1. Username and password: set both Username and Password fields, and leave the Pat
//     field empty.
//
//  2. Personal access token: set the Username and Pat fields, and leave the Password
//     field empty.
//
// Pat field has more priority than Password field, meaning that if both are set, the
// Pat field will be used.
// TODO: support key authentication
type GitAuth struct {
	Username string
	Password string
	Pat      string
}

func (g *NewPackageHandlerOptions) getAuth() *http.BasicAuth {
	if g.GitAuth == nil {
		return nil
	}
	if g.GitAuth.Pat != "" {
		return &http.BasicAuth{
			Username: g.GitAuth.Username,
			Password: g.GitAuth.Pat,
		}
	}
	return &http.BasicAuth{
		Username: g.GitAuth.Username,
		Password: g.GitAuth.Password,
	}
}

// NewPackageHandlerFromURL clones the package from the given URL and returns. The GitAuth
// field could be used to provide authentication to a private git repository.
func NewPackageHandlerFromURL(opts NewPackageHandlerOptions) (*PackageHandler, error) {
	_, err := git.PlainClone(opts.Path, false, &git.CloneOptions{
		URL:  opts.URL,
		Auth: opts.getAuth(),
	})
	if err != nil {
		if errors.Is(err, transport.ErrAuthenticationRequired) {
			return nil, RepositoryNotFoundOrPrivateError{
				URL: opts.URL,
			}
		}
		if errors.Is(err, transport.ErrRepositoryNotFound) {
			return nil, RepositoryNotFoundError{
				URL: opts.URL,
			}
		}
		return nil, err
	}
	return NewPackageHandler(opts.Path), nil
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
	}
	return p.checkSum()
}

// Versions returns the descending sorted list of available versions for the package.
// A version is a git tag that matches the regex `^v\d+\.\d+\.\d+$`.
func (p *PackageHandler) Versions() ([]string, error) {
	pkgRepo, err := git.PlainOpen(p.path)
	if err != nil {
		return nil, err
	}
	tagIter, err := pkgRepo.Tags()
	if err != nil {
		return nil, err
	}
	var versions []string
	tagIter.ForEach(func(ref *plumbing.Reference) error {
		tag := ref.Name().Short()
		if tagVersionRegex.MatchString(tag) {
			versions = append(versions, tag)
		}
		return nil
	})
	if len(versions) == 0 {
		return nil, ErrNoVersionsFound
	}
	sort.Slice(versions, func(i, j int) bool {
		return strings.ToLower(versions[i]) > strings.ToLower(versions[j])
	})
	return versions, nil
}

// HasVersion returns an error if the given version is not available for the package.
func (p *PackageHandler) HasVersion(version string) error {
	versions, err := p.Versions()
	if err != nil {
		return err
	}
	for _, v := range versions {
		if v == version {
			return nil
		}
	}
	return fmt.Errorf("%w: %s", ErrVersionNotFound, version)
}

// LatestVersion returns the latest version of the package.
func (p *PackageHandler) LatestVersion() (string, error) {
	versions, err := p.Versions()
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

// CheckoutVersion checks out the cloned repository to the given version (tag).
func (p *PackageHandler) CheckoutVersion(version string) error {
	if !tagVersionRegex.MatchString(version) {
		return ErrInvalidVersion
	}
	gitRepo, err := git.PlainOpen(p.path)
	if err != nil {
		return err
	}
	tagIter, err := gitRepo.Tags()
	if err != nil {
		return err
	}
	defer tagIter.Close()
	for {
		tag, err := tagIter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return ErrNoVersionsFound
			}
			return err
		}
		if tag.Name().Short() == version {
			worktree, err := gitRepo.Worktree()
			if err != nil {
				return fmt.Errorf("error getting worktree: %w", err)
			}
			err = worktree.Checkout(&git.CheckoutOptions{
				Branch: tag.Name(),
			})
			if err != nil {
				return err
			}
			break
		}
	}
	return nil
}

// CurrentVersion returns the current version of the package, which is tha latest
// tag with version format that points to the current HEAD.
func (p *PackageHandler) CurrentVersion() (string, error) {
	gitRepo, err := git.PlainOpen(p.path)
	if err != nil {
		return "", err
	}
	head, err := gitRepo.Head()
	if err != nil {
		return "", err
	}
	tagIter, err := gitRepo.TagObjects()
	if err != nil {
		return "", err
	}
	var headVersions []string
	for {
		tag, err := tagIter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if tagVersionRegex.MatchString(tag.Name) && head.Hash() == tag.Target {
			headVersions = append(headVersions, tag.Name)
		}
	}
	if len(headVersions) == 0 {
		return "", ErrNoVersionsFound
	}
	sort.Slice(headVersions, func(i, j int) bool {
		return strings.ToLower(headVersions[i]) > strings.ToLower(headVersions[j])
	})
	return headVersions[0], nil
}

// Profiles returns the list of profiles defined in the package for the current version.
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
		profile.Name = profileName

		if err := profile.Validate(); err != nil {
			return nil, err
		}

		profiles = append(profiles, *profile)
	}

	return profiles, nil
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
		env[strings.Trim(parts[0], " ")] = strings.Trim(parts[1], " ")
	}
	return env, nil
}

// ProfileFS returns the filesystem for the given profile.
func (p *PackageHandler) ProfileFS(profileName string) fs.FS {
	return os.DirFS(filepath.Join(p.path, pkgDirName, profileName))
}

// HasPlugin returns true if the package has a plugin.
func (p *PackageHandler) HasPlugin() (bool, error) {
	manifest, err := p.parseManifest()
	if err != nil {
		return false, err
	}

	return manifest.Plugin != nil, nil
}

// Plugin returns the plugin for the package.
func (p *PackageHandler) Plugin() (*Plugin, error) {
	manifest, err := p.parseManifest()
	if err != nil {
		return nil, err
	}

	if manifest.Plugin == nil {
		return nil, ErrNoPlugin
	}

	return manifest.Plugin, nil
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
