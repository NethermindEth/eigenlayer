package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"path"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/internal/locker"
	"github.com/NethermindEth/egn/internal/package_handler"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Checks that WizDaemon implements Daemon.
var _ = Daemon(&WizDaemon{})

// WizDaemon is the main entrypoint for all the functionalities of the daemon.
type WizDaemon struct {
	dataDir       *data.DataDir
	dockerCompose ComposeManager
	monitoringMgr MonitoringManager
	fs            afero.Fs
	locker        locker.Locker
}

// NewDaemon create a new daemon instance.
func NewWizDaemon(
	cmpMgr ComposeManager,
	mtrMgr MonitoringManager,
	fs afero.Fs,
	locker locker.Locker,
) (*WizDaemon, error) {
	dataDir, err := data.NewDataDirDefault(fs, locker)
	if err != nil {
		return nil, err
	}
	return &WizDaemon{
		dataDir:       dataDir,
		dockerCompose: cmpMgr,
		monitoringMgr: mtrMgr,
		fs:            fs,
		locker:        locker,
	}, nil
}

// Init initializes the daemon.
func (d *WizDaemon) Init() error {
	// *** Monitoring stack initialization. ***
	// Check if the monitoring stack is installed.
	installStatus, err := d.monitoringMgr.InstallationStatus()
	if err != nil {
		return err
	}
	log.Infof("Monitoring stack installation status: %v", installStatus == common.Installed)
	// If the monitoring stack is not installed, install it.
	if installStatus == common.NotInstalled {
		return d.monitoringMgr.InitStack()
	}
	// Check if the monitoring stack is running.
	status, err := d.monitoringMgr.Status()
	if err != nil {
		return err
	}
	// If the monitoring stack is not running, start it.
	if status != common.Running && status != common.Restarting {
		if err := d.monitoringMgr.Run(); err != nil {
			return err
		}
	}
	// *** Monitoring stack initialization. ***
	return nil
}

// Pull implements Daemon.Pull.
func (d *WizDaemon) Pull(url string, version string, force bool) (result PullResult, err error) {
	tID := tempID(url)
	if force {
		if err = d.dataDir.RemoveTemp(tID); err != nil {
			return
		}
	}
	tempPath, err := d.dataDir.InitTemp(tID)
	if err != nil {
		return
	}
	pkgHandler, err := package_handler.NewPackageHandlerFromURL(package_handler.NewPackageHandlerOptions{
		Path: tempPath,
		URL:  url,
	})
	if err != nil {
		return
	}
	if err = pkgHandler.Check(); err != nil {
		return
	}
	// Set version
	if version == "" {
		version, err = pkgHandler.LatestVersion()
		if err != nil {
			return
		}
	}
	result.Version = version
	// Get profiles names and its options
	profiles, err := pkgHandler.Profiles()
	if err != nil {
		return
	}
	profileOptions := make(map[string][]Option, len(profiles))
	for _, profile := range profiles {
		options := make([]Option, len(profile.Options))
		for i, o := range profile.Options {
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
			case "select":
				options[i] = NewOptionSelect(o)
			case "port":
				options[i], err = NewOptionPort(o)
			case "id":
				options[i] = NewOptionID(o)
			default:
				err = errors.New("unknown option type: " + o.Type)
				return
			}
		}
		if err != nil {
			return
		}
		profileOptions[profile.Name] = options
	}
	result.Options = profileOptions
	return result, nil
}

// Install implements Daemon.Install.
func (d *WizDaemon) Install(options InstallOptions) (string, error) {
	// Get pulled package directory from temp
	tID := tempID(options.URL)
	tempPath, err := d.dataDir.TempPath(tID)
	if err != nil {
		return "", err
	}

	instanceName, err := instanceNameFromURL(options.URL)
	if err != nil {
		return "", err
	}
	instanceId := data.InstanceId(instanceName, options.Tag)

	// Check if instance already exists
	if d.dataDir.HasInstance(instanceId) {
		return "", fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceId)
	}

	// Init package handler from temp path
	pkgHandler := package_handler.NewPackageHandler(tempPath)
	// Check if selected version is valid
	if err := pkgHandler.HasVersion(options.Version); err != nil {
		return "", err
	}
	if err = pkgHandler.CheckoutVersion(options.Version); err != nil {
		return "", err
	}

	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return "", err
	}
	var selectedProfile *package_handler.Profile
	// Check if selected profile is valid
	for _, pkgProfile := range pkgProfiles {
		if pkgProfile.Name == options.Profile {
			selectedProfile = &pkgProfile
			break
		}
	}
	if selectedProfile == nil {
		return "", fmt.Errorf("%w: %s", ErrProfileDoesNotExist, options.Profile)
	}

	// Install package
	env := make(map[string]string, len(options.Options))
	for _, o := range options.Options {
		env[o.Target()] = o.Value()
	}
	instance := data.Instance{
		Name:    instanceName,
		Profile: selectedProfile.Name,
		Version: options.Version,
		URL:     options.URL,
	}
	err = d.dataDir.InitInstance(&instance)
	if err != nil {
		return "", err
	}
	return instanceId, instance.Setup(env, pkgHandler.ProfileFS(instance.Profile))
}

// Run implements Daemon.Run.
func (d *WizDaemon) Run(instanceID string) error {
	instancePath, err := d.dataDir.InstancePath(instanceID)
	if err != nil {
		return err
	}
	composePath := path.Join(instancePath, "docker-compose.yml")
	return d.dockerCompose.Up(compose.DockerComposeUpOptions{
		Path: composePath,
	})
}

func instanceNameFromURL(u string) (string, error) {
	parsedURL, err := url.ParseRequestURI(u)
	if err != nil {
		return "", err
	}
	return path.Base(parsedURL.Path), nil
}

func tempID(url string) string {
	tempHash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(tempHash[:])
}
