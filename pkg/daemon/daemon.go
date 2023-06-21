package daemon

/* TODO: The Install feature could be split into multiple steps:
1. 	Pull the package to a temporary directory inside the data directory.
2. 	Ask the user to select a profile from the available profiles in the package
	and fill all the options.
3. 	Copy the selected profile and the .env resulting from the options filling
    to the instances directory.
4. 	Run the instance if the user confirms it.

This way, the daemon will have multiple methods that can be used together to
implement the Install feature. Those methods will be:

- Pull: clone the package to a temporary directory inside the data directory,
		the `tmp` directory at the root of the data directory. Each pull stores
		the package in a directory with the package hash as the name to facilitate
		the cache of the packages.
- Install: checks if the package is already pulled, searching by the package hash
		on the `tmp` directory. If it is not pulled returns an error, otherwise, it asks
		the user to select a profile and fill the options. Then, it copies the selected
		profile and the .env file resulting from the options filling to the instances directory.
		At the end of the installation removes the package from the `tmp` directory.
- Run: checks if the instance is already installed on the instances directory. If
		it is not installed returns an error, otherwise, it runs the docker-compose up -d
		command on the instance directory.*/

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
