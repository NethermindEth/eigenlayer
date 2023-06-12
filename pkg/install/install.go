package install

import (
	"errors"
	"fmt"
	"io"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Installer is used to install AVS node software from a git repository.
type Installer struct {
	gitAuth *GitAuth
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

// NewInstaller returns a new Installer instance. If you need to install from a private
// repository, use NewInstallerWithAuth instead.
func NewInstaller() *Installer {
	return &Installer{gitAuth: nil}
}

// NewInstallerWithAuth returns a new Installer instance with git credentials. If you
// need to install from a public repository, NewInstaller will be sufficient.
func NewInstallerWithAuth(gitAuth GitAuth) *Installer {
	return &Installer{
		gitAuth: &gitAuth,
	}
}

// Install installs the AVS node software from the given git repository URL and version
// to the given destination directory.
func (i *Installer) Install(url, version, dest string) error {
	if err := cloneGitRepo(url, version, dest, i.getAuth()); err != nil {
		return err
	}

	pkgHandler := package_handler.NewPackageHandler(dest)
	return pkgHandler.Check()
}

func (g *Installer) getAuth() *http.BasicAuth {
	if g.gitAuth == nil {
		return nil
	}
	if g.gitAuth.Pat != "" {
		return &http.BasicAuth{
			Username: g.gitAuth.Username,
			Password: g.gitAuth.Pat,
		}
	}
	return &http.BasicAuth{
		Username: g.gitAuth.Username,
		Password: g.gitAuth.Password,
	}
}

func cloneGitRepo(url, tagName, dest string, auth *http.BasicAuth) error {
	gitRepo, err := git.PlainClone(dest, false, &git.CloneOptions{
		URL:  url,
		Auth: auth,
	})
	if err != nil {
		if errors.Is(err, transport.ErrAuthenticationRequired) {
			return RepositoryNotFoundOrPrivateError{
				URL: url,
			}
		}
		if errors.Is(err, transport.ErrRepositoryNotFound) {
			return RepositoryNotFoundError{
				URL: url,
			}
		}
		return err
	}

	tag, err := getTag(gitRepo, tagName)
	if err != nil {
		return err
	}
	worktree, err := gitRepo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}
	return worktree.Checkout(&git.CheckoutOptions{
		Branch: tag.Name(),
	})
}

func getTag(gitRepo *git.Repository, tag string) (*plumbing.Reference, error) {
	tagsIter, err := gitRepo.Tags()
	if err != nil {
		return nil, fmt.Errorf("error getting tags: %w", err)
	}
	for {
		n, err := tagsIter.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("error getting next tag: %w", err)
		}
		if n.Name().Short() == tag {
			tagsIter.Close()
			return n, nil
		}
	}
	return nil, TagNotFoundError{Tag: tag}
}
