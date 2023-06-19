package daemon

// Daemon is the interface for the egn daemon. It should be used as the entrypoint
// for all the functionalities of egn.
type Daemon interface {
	// Install downloads and installs a node software package using the provided options.
	Install(options InstallOptions) error
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	// URL is the URL of the git repository containing the node software package.
	URL string

	// Version is the version of the node software to install. If empty, the latest
	// version will be installed. A package version is a git tag that matches the
	// regex `^v\d+\.\d+\.\d+$`.
	Version string

	// Tag is the tag to use for the instance. If empty, the `default` tag will
	// be used. Tag is used to differentiate between multiple instances of the
	// same package name.
	Tag string

	// ProfileSelector is used by the daemon to ask the user to select a profile
	// from the available profiles in the package.
	ProfileSelector func(profiles []string) (string, error)

	// OptionsFiller is used by the daemon to ask the user to fill the options
	// for the selected profile.
	OptionsFiller func(opts []Option) ([]Option, error)

	// RunConfirmation is used by the daemon to ask the user to confirm the run
	// of the instance after the installation.
	RunConfirmation func() (bool, error)
}
