package daemon

import (
	"errors"

	"github.com/NethermindEth/eigen-wiz/pkg/pull"
)

// Checks that WizDaemon implements Daemon.
var _ = Daemon(&WizDaemon{})

// WizDaemon is the main entrypoint for all the functionalities of the daemon.
type WizDaemon struct{}

// NewDaemon create a new daemon instance.
func NewWizDaemon() *WizDaemon {
	return &WizDaemon{}
}

type PullOptions struct {
	URL     string
	Version string
	DestDir string
}

type PullResponse struct {
	CurrentVersion string
	LatestVersion  string
	Profiles       []Profile
}
type Profile struct {
	Name    string
	Options []string
}

func (d *WizDaemon) Pull(options *PullOptions) (*PullResponse, error) {
	pkgHandler, err := pull.Pull(options.URL, options.Version, options.DestDir)
	if err != nil {
		return nil, err
	}
	currentVersion, err := pkgHandler.CurrentVersion()
	if err != nil {
		return nil, err
	}
	latestVersion, err := pkgHandler.CurrentVersion()
	if err != nil {
		return nil, err
	}
	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return nil, err
	}
	var profiles []Profile
	for _, pkgProfile := range pkgProfiles {
		options := make([]string, len(pkgProfile.Options))
		for i, option := range pkgProfile.Options {
			options[i] = option.Name
		}
		profiles = append(profiles, Profile{
			Name:    pkgProfile.Name,
			Options: options,
		})
	}
	return &PullResponse{
		CurrentVersion: currentVersion,
		LatestVersion:  latestVersion,
		Profiles:       profiles,
	}, nil
}

type RunOptions struct{}

type RunResponse struct{}

func (d *WizDaemon) Run(options *RunOptions) (*RunResponse, error) {
	return &RunResponse{}, errors.New("not implemented")
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	PullOptions
}

// InstallResponse is a response from installing a node software package.
type InstallResponse struct{}

// Install installs a node software package using the provided options.
func (d *WizDaemon) Install(options *InstallOptions) (*InstallResponse, error) {
	_, err := pull.Pull(options.URL, options.Version, options.DestDir)
	if err != nil {
		return nil, err
	}
	// TODO: run package from the pkgHandler
	return &InstallResponse{}, nil
}
