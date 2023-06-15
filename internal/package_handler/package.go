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

type NewPackageHandlerOptions struct {
	Path    string
	URL     string
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
	sort.Slice(versions, func(i, j int) bool {
		return strings.ToLower(versions[i]) > strings.ToLower(versions[j])
	})
	return versions, nil
}

func (p *PackageHandler) LatestVersion() (string, error) {
	versions, err := p.Versions()
	if err != nil {
		return "", err
	}
	return versions[0], nil
}

func (p *PackageHandler) CheckoutVersion(version string) error {
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
				return fmt.Errorf("version %s not found", version)
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
	for {
		tag, err := tagIter.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}
		if head.Hash() == tag.Target {
			return tag.Name, nil
		}
	}
	return "", errors.New("no tag found for current version")
}

func (p *PackageHandler) Run() error {
	return errors.New("not implemented")
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
		profile.Name = profileName

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
		env[strings.Trim(parts[0], " ")] = strings.Trim(parts[1], " ")
	}
	return env, nil
}

func (p *PackageHandler) ProfileFS(profileName string) fs.FS {
	return os.DirFS(filepath.Join(p.path, pkgDirName, profileName))
}
