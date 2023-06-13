package pull

import (
	"errors"
	"regexp"

	"github.com/NethermindEth/eigen-wiz/internal/package_handler"
)

var tagVersionRegex = regexp.MustCompile(`^v\d+\.\d+\.\d+$`)

// Pull downloads the AVS node software from the given git repository URL and version to
// the given destination directory. If the version is empty, the latest version will be
// downloaded, but if the repo doesn't have any tags that match the version format, an
// error will be returned.
func Pull(url, version, dest string) (*package_handler.PackageHandler, error) {
	if !tagVersionRegex.MatchString(version) {
		return nil, errors.New("invalid version format")
	}

	pkgHandler, err := package_handler.NewPackageHandlerFromURL(package_handler.NewPackageHandlerOptions{
		Path: dest,
		URL:  url,
	})
	if err != nil {
		return nil, err
	}

	if version == "" {
		version, err = pkgHandler.LatestVersion()
		if err != nil {
			return nil, err
		}
	}

	if err = pkgHandler.CheckoutVersion(version); err != nil {
		return nil, err
	}

	if err = pkgHandler.Check(); err != nil {
		return nil, err
	}

	return pkgHandler, nil
}
