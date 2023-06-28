package daemon

// Daemon is the interface for the egn daemon. It should be used as the entrypoint
// for all the functionalities of egn.
type Daemon interface {
	// Pull downloads a node software package from the given URL and returns the
	// version and options of each profile in the package. If force is true and
	// the package already exists, it will be removed and re-downloaded. After
	// calling Pull all is ready to call Install.
	Pull(url string, version string, force bool) (PullResult, error)

	// Install downloads and installs a node software package using the provided options,
	// and returns the instance ID of the installed package. Make sure to call Pull
	// before calling Install to ensure that the package is downloaded.
	Install(options InstallOptions) (string, error)

	// Run starts the instance with the given ID running docker compose in the
	// instance directory. If there is no installed instance with the given ID,
	// an error will be returned.
	Run(instanceId string) error

	// Stop stops the instance with the given ID. If there is no installed instance
	// with the given ID an error will be returned.
	Stop(instanceId string) error

	// Uninstall stops and removes the instance with the given ID. If there is no
	// installed instance with the given ID an error will be returned.
	Uninstall(instanceId string) error

	// Init initializes the daemon, making sure that all the necessary components
	// are installed and running.
	Init() error
}

// PullResult is the result of a Pull operation, containing all the necessary
// information from the package.
type PullResult struct {
	// Version is the version of the pulled package.
	Version string

	// Options is map of profile names to their options.
	Options map[string][]Option
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	// URL is the URL of the git repository containing the node software package.
	URL string

	// Version is the version of the node software to install. If empty, the latest
	// version will be installed. A package version is a git tag that matches the
	// regex `^v\d+\.\d+\.\d+$`.
	Version string

	// Profile is the name of the profile to use for the instance.
	Profile string

	// Options is the list of options to use for the instance.
	Options []Option

	// Tag is the tag to use for the instance, required to build the instance id
	// with the format <package_name>-<tag>
	Tag string
}
