package daemon

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Daemon is the interface for the egn daemon. It should be used as the entrypoint
// for all the functionalities of egn.
type Daemon interface {
	// Pull downloads a node software package from the given URL and returns the
	// version and options of each profile in the package. If force is true and
	// the package already exists, it will be removed and re-downloaded. After
	// calling Pull all is ready to call Install.
	Pull(url string, ref PullTarget, force bool) (PullResult, error)

	// PullUpdate downloads a node software package from the given URL and returns
	// the result of merging both packages configs.
	PullUpdate(instanceID string, ref PullTarget) (PullUpdateResult, error)

	// LocalPullUpdate loads a node software package from a local tarball and
	// returns the result of merging both packages configs.
	LocalPullUpdate(instanceID string, pkgTar io.Reader) (PullUpdateResult, error)

	// Install downloads and installs a node software package using the provided options,
	// and returns the instance ID of the installed package. Make sure to call Pull
	// before calling Install to ensure that the package is downloaded.
	Install(options InstallOptions) (string, error)

	// HasInstance returns true if there is an installed instance with the given ID.
	HasInstance(instanceId string) bool

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

	// InitMonitoring initializes the MonitoringStack. If install is true, the
	// MonitoringStack will be installed if it is not already installed. If run
	// is true, the MonitoringStack will be run if it is not already running.
	InitMonitoring(install, run bool) error

	// CleanMonitoring stops and uninstalls the MonitoringStack
	CleanMonitoring() error

	// RunPlugin runs a plugin with the given arguments on the instance with the
	// given ID. If there is no installed and running instance with the given ID
	// an error will be returned. If noDestroyImage is true, the plugin image will
	// not be removed after the plugin execution.
	RunPlugin(instanceId string, pluginArgs []string, options RunPluginOptions) error

	// CheckHardwareRequirements checks if the hardware of the system meets the
	// specified requirements. It takes a HardwareRequirements struct as input and returns
	// a boolean value indicating whether the hardware meets the requirements.
	CheckHardwareRequirements(requirements HardwareRequirements) (bool, error)

	// ListInstances returns a list of all the installed instances and their health.
	ListInstances() ([]ListInstanceItem, error)

	// LocalInstall installs a node software package from a local tarball. This
	// installation method is only intended for development purposes and is not
	// secure. It returns the instance ID of the installed package.
	LocalInstall(pkgTar io.Reader, options LocalInstallOptions) (string, error)

	// NodeLogs returns the logs of the node with the given ID. If there is no
	// installed instance with the given ID an error will be returned.
	NodeLogs(ctx context.Context, w io.Writer, instanceID string, opts NodeLogsOptions) error

	// Backup creates a backup of the instance with the given ID and returns the
	// backup ID. If there is no installed instance with the given ID an error
	// will be returned.
	Backup(instanceId string) (backupId string, err error)

	// Restore restores the backup with the given ID. If the AVS instance id of
	// the backup exists, then the command will uninstall it before restoring
	// the backup. If the AVS instance does not exist, then the command will
	// create it. If run is true, the instance will be run after the restore.
	Restore(backupId string, run bool) error

	// BackupList returns a list of all the backups and their information.
	BackupList() ([]BackupInfo, error)
}

type PullTarget struct {
	Version string
	Commit  string
}

type RunPluginOptions struct {
	NoDestroyImage bool
	HostNetwork    bool
	Binds          map[string]string
	Volumes        map[string]string
}

// ListInstanceItem is an item in the list of instances returned by ListInstances.
type ListInstanceItem struct {
	ID      string
	Version string
	Commit  string
	Health  NodeHealth
	Running bool
	Comment string
}

// NodeHealth is the health of a node, matching the HTTP status codes.
type NodeHealth int

const (
	NodeHealthUnknown    NodeHealth = 0
	NodeHealthy          NodeHealth = 200
	NodePartiallyHealthy NodeHealth = 206
	NodeUnhealthy        NodeHealth = 503
)

func (n NodeHealth) String() string {
	switch n {
	case NodeHealthy:
		return "healthy"
	case NodePartiallyHealthy:
		return "partially healthy"
	case NodeUnhealthy:
		return "unhealthy"
	default:
		return "unknown"
	}
}

type NodeLogsOptions struct {
	Follow     bool
	Since      string
	Until      string
	Timestamps bool
	Tail       string
}

// PullResult is the result of a Pull operation, containing all the necessary
// information from the package.
type PullResult struct {
	// Name is the name of the AVS represented by the pulled package.
	Name string

	// Version is the version of the pulled package.
	Version string

	// SpecVersion is the version of the Eigenlayer AVS Node Specification the instance
	// targets. The version must match the regex `^\d+\.\d+\.\d+$`.
	SpecVersion string

	// Commit hash of the pulled package.
	Commit string

	// HasPlugin is true if the package has a plugin.
	HasPlugin bool

	// Options is map of profile names to their options.
	Options map[string][]Option

	// HardwareRequirements is the hardware requirements specified in the package manifest.
	HardwareRequirements map[string]HardwareRequirements
}

type PullUpdateResult struct {
	Name    string
	Tag     string
	Url     string
	Profile string
	// OldVersion is the version of the old package.
	OldVersion string

	// NweVersion is the version of the new package.
	NewVersion string

	// OldCommit is the commit hash of the old package.
	OldCommit string

	// NewCommit is the commit hash of the new package.
	NewCommit string

	// HasPlugin is true if the package has a plugin.
	HasPlugin bool

	// OldOptions is the list of options of the old package.
	OldOptions []Option

	// NewOptions is the list of options of the new package.
	NewOptions []Option

	// MergedOptions is the list of options of the new package merged with the
	// old package. These options are the ones that will be used for the new instance
	// and should be filled by the user.
	MergedOptions []Option

	// HardwareRequirements is the hardware requirements specified in the package manifest.
	HardwareRequirements HardwareRequirements
}

// InstallOptions is a set of options for installing a node software package.
type InstallOptions struct {
	// Name is the name of the AVS represented by the package.
	Name string

	// Tag is the tag to use for the instance, required to build the instance id
	// with the format <package_name>-<tag>
	Tag string

	// URL is the URL of the git repository containing the node software package.
	URL string

	// Version is the version of the node software to install. If empty, the latest
	// version will be installed. A package version is a git tag that matches the
	// regex `^v\d+\.\d+\.\d+$`.
	Version string

	// SpecVersion is the version of the Eigenlayer AVS Node Specification the instance
	// targets. The version must match the regex `^\d+\.\d+\.\d+$`.
	SpecVersion string

	// Commit is the commit to install from. It has precedence over Version.
	Commit string

	// Profile is the name of the profile to use for the instance.
	Profile string

	// Options is the list of options to use for the instance.
	Options []Option
}

// LocalInstallOptions is a set of options for installing a node software package
// from a local tarball.
type LocalInstallOptions struct {
	// Name is the name of the package.
	Name string

	// Tag is the tag to use for the instance, required to build the instance id
	// with the format <package_name>-<tag>
	Tag string

	// Profile is the name of the profile to use for the instance.
	Profile string

	// Options is the list of options to use for the instance. These options are
	// passed as strings because the local installation method is for development
	// purposes only, and the user is responsible for passing the correct options.
	Options map[string]string
}

type HardwareRequirements struct {
	MinCPUCores                 int
	MinRAM                      int
	MinFreeSpace                int
	StopIfRequirementsAreNotMet bool
}

func (h HardwareRequirements) String() string {
	return fmt.Sprintf("CPU: %d Cores, RAM: %d Mb, Disk Space: %d Mb", h.MinCPUCores, h.MinRAM, h.MinFreeSpace)
}

type BackupInfo struct {
	Id        string
	Instance  string
	Timestamp time.Time
	SizeBytes int64
	Version   string
	Commit    string
	Url       string
}
