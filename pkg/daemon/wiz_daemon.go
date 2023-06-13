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
	Profiles       map[string][]Option
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
	profiles := make(map[string][]Option, len(pkgProfiles))
	for _, pkgProfile := range pkgProfiles {
		options := make([]Option, len(pkgProfile.Options))
		for i, o := range pkgProfile.Options {
			switch o.Type {
			case "str":
				options[i] = NewOptionString(o)
			case "int":
				options[i], err = NewOptionInt(o)
			case "float":
				options[i], err = NewOptionFloat(o)
			case "bool":
				options[i], err = NewOptionBool(o)
			case "path_dir":
				options[i] = NewOptionPathDir(o)
			case "path_file":
				options[i] = NewOptionPathFile(o)
			case "uri":
				options[i] = NewOptionURI(o)
			case "enum":
				options[i] = NewOptionEnum(o)
			case "port":
				options[i], err = NewOptionPort(o)
			case "id":
				options[i] = NewOptionID(o)
			default:
				return nil, errors.New("unknown option type: " + o.Type)
			}
		}
		if err != nil {
			return nil, err
		}
		profiles[pkgProfile.Name] = options
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
