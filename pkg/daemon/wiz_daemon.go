package daemon

// Installer is an interface for installing a node software package.
type Installer interface {
	Install(url, version, destDir string) error
}

// Checks that WizDaemon implements Daemon.
var _ = Daemon(&WizDaemon{})

// WizDaemon is the main entrypoint for all the functionalities of the daemon.
type WizDaemon struct {
	installer Installer
}

// NewDaemon create a new daemon instance.
func NewWizDaemon(installer Installer) *WizDaemon {
	return &WizDaemon{
		installer: installer,
	}
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	URL     string
	Version string
	DestDir string
}

// InstallResponse is a response from installing a node software package.
type InstallResponse struct{}

// Install installs a node software package using the provided options.
func (d *WizDaemon) Install(options *InstallOptions) (*InstallResponse, error) {
	if err := d.installer.Install(options.URL, options.Version, options.DestDir); err != nil {
		return nil, err
	}
	return &InstallResponse{}, nil
}
