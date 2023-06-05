package daemon

// Installer is an interface for installing a node software package.
type Installer interface {
	Install(url, version, destDir string) error
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
func (d *Daemon) Install(options *InstallOptions) (*InstallResponse, error) {
	if err := d.installer.Install(options.URL, options.Version, options.DestDir); err != nil {
		return nil, err
	}
	return &InstallResponse{}, nil
}
