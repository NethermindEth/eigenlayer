package daemon

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"path"
	"strconv"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/internal/locker"
	"github.com/NethermindEth/egn/internal/monitoring"
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
	docker        DockerManager
	monitoringMgr MonitoringManager
	fs            afero.Fs
	locker        locker.Locker
}

// NewDaemon create a new daemon instance.
func NewWizDaemon(
	cmpMgr ComposeManager,
	dockerMgr DockerManager,
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
		docker:        dockerMgr,
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
		err = d.monitoringMgr.InstallStack()
		if errors.Is(err, monitoring.ErrInstallingMonitoringMngr) {
			// If the monitoring stack installation fails, remove the monitoring stack directory.
			if cerr := d.monitoringMgr.Cleanup(true); cerr != nil {
				return fmt.Errorf("install failed: %w. Failed to cleanup monitoring stack after installation failure: %w", err, cerr)
			}
			return err
		} else if err != nil {
			return err
		}
		return nil
	}
	// Check if the monitoring stack is running.
	status, err := d.monitoringMgr.Status()
	if err != nil {
		log.Errorf("Monitoring stack status: unknown. Got error: %v", err)
	}
	// If the monitoring stack is not running, start it.
	if status != common.Running && status != common.Restarting {
		if err := d.monitoringMgr.Run(); err != nil {
			return err
		}
	}
	if err := d.monitoringMgr.Init(); err != nil {
		return err
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
	instanceId, tempDirID, err := d.install(options)
	if err != nil && instanceId != "" {
		// Cleanup if Install fails
		if cerr := d.uninstall(instanceId, false); cerr != nil {
			err = fmt.Errorf("install failed: %w. Failed to cleanup after installation failure: %w", err, cerr)
		}
	}

	// Cleanup temp folder
	if rerr := d.dataDir.RemoveTempDir(tempDirID); rerr != nil {
		if err != nil {
			err = fmt.Errorf("install failed: %w. Failed to cleanup temporary folder after installation failure: %w", err, rerr)
		} else {
			err = fmt.Errorf("install failed. Failed to cleanup temporary folder after installation failure: %w", rerr)
		}
	}
	return instanceId, err
}

func (d *WizDaemon) install(options InstallOptions) (string, string, error) {
	// Get pulled package directory from temp
	tID := tempID(options.URL)
	tempPath, err := d.dataDir.TempPath(tID)
	if err != nil {
		return "", tID, err
	}

	instanceName, err := instanceNameFromURL(options.URL)
	if err != nil {
		return "", tID, err
	}
	instanceId := data.InstanceId(instanceName, options.Tag)

	// Check if instance already exists
	if d.dataDir.HasInstance(instanceId) {
		return "", tID, fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceId)
	}

	// Init package handler from temp path
	pkgHandler := package_handler.NewPackageHandler(tempPath)
	// Check if selected version is valid
	if err := pkgHandler.HasVersion(options.Version); err != nil {
		return "", tID, err
	}
	if err = pkgHandler.CheckoutVersion(options.Version); err != nil {
		return "", tID, err
	}

	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return "", tID, err
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
		return "", tID, fmt.Errorf("%w: %s", ErrProfileDoesNotExist, options.Profile)
	}

	// Install package
	env := make(map[string]string, len(options.Options))
	for _, o := range options.Options {
		env[o.Target()] = o.Value()
	}
	// Get monitoring targets
	monitoringTargets := make([]data.MonitoringTarget, 0)
	for _, target := range selectedProfile.Monitoring.Targets {
		mt := data.MonitoringTarget{
			Service: target.Service,
			Port:    strconv.Itoa(target.Port),
			Path:    target.Path,
		}
		monitoringTargets = append(monitoringTargets, mt)
	}

	instance := data.Instance{
		Name:              instanceName,
		Profile:           selectedProfile.Name,
		Version:           options.Version,
		URL:               options.URL,
		Tag:               options.Tag,
		MonitoringTargets: data.MonitoringTargets{Targets: monitoringTargets},
	}
	if err = d.dataDir.InitInstance(&instance); err != nil {
		return instanceId, tID, err
	}

	if err = instance.Setup(env, pkgHandler.ProfileFS(instance.Profile)); err != nil {
		return instanceId, tID, err
	}

	// Create containers
	// TODO: Log Create output and log to wait as containers might be built
	if err = d.dockerCompose.Create(compose.DockerComposeCreateOptions{
		Path:  instance.ComposePath(),
		Build: true,
	}); err != nil {
		return instanceId, tID, err
	}

	// Start containers
	if err = d.dockerCompose.Up(compose.DockerComposeUpOptions{
		Path: instance.ComposePath(),
	}); err != nil {
		return instanceId, tID, err
	}

	if err = d.addTarget(instanceId); err != nil {
		return instanceId, tID, err
	}

	return instanceId, tID, nil
}

// Run implements Daemon.Run.
func (d *WizDaemon) Run(instanceID string) error {
	// Add target just in case
	if err := d.addTarget(instanceID); err != nil {
		return err
	}

	instancePath, err := d.dataDir.InstancePath(instanceID)
	if err != nil {
		return err
	}
	composePath := path.Join(instancePath, "docker-compose.yml")
	return d.dockerCompose.Up(compose.DockerComposeUpOptions{
		Path: composePath,
	})
}

// Stop implements Daemon.Stop.
func (d *WizDaemon) Stop(instanceID string) error {
	instancePath, err := d.dataDir.InstancePath(instanceID)
	if err != nil {
		return err
	}
	composePath := path.Join(instancePath, "docker-compose.yml")
	return d.dockerCompose.Stop(compose.DockerComposeStopOptions{
		Path: composePath,
	})
}

// Uninstall implements Daemon.Uninstall.
func (d *WizDaemon) Uninstall(instanceID string) error {
	return d.uninstall(instanceID, true)
}

func (d *WizDaemon) uninstall(instanceID string, down bool) error {
	if err := d.removeTarget(instanceID); err != nil {
		return err
	}

	if down {
		instancePath, err := d.dataDir.InstancePath(instanceID)
		if err != nil {
			return err
		}
		composePath := path.Join(instancePath, "docker-compose.yml")
		// docker compose down
		if err = d.dockerCompose.Down(compose.DockerComposeDownOptions{
			Path: composePath,
		}); err != nil {
			return err
		}
	}

	// remove instance directory
	return d.dataDir.RemoveInstance(instanceID)
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

type psServiceJSON struct {
	ID      string `json:"ID"`
	Service string `json:"Service"`
}

func (d *WizDaemon) monitoringTargetsEndpoints(serviceNames []string, composePath string) (map[string]string, error) {
	psOut, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:   composePath,
		Format: "json",
		All:    true,
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal docker-compose ps output
	var psServices []psServiceJSON
	if err = json.Unmarshal([]byte(psOut), &psServices); err != nil {
		return nil, err
	}

	// Get containerID of monitoring targets
	monitoringTargets := make(map[string]string)
	for _, serviceName := range serviceNames {
		for _, psService := range psServices {
			if psService.Service == serviceName {
				monitoringTargets[serviceName] = psService.ID
			}
		}
	}

	return monitoringTargets, nil
}

func (d *WizDaemon) idToEndpoint(id, path, port string) (string, error) {
	ip, err := d.docker.ContainerIP(id)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://%s:%s", ip, port), nil
}

func (d *WizDaemon) addTarget(instanceID string) error {
	// Get monitoring targets
	instance, err := d.dataDir.Instance(instanceID)
	if err != nil {
		return err
	}
	// Get containerID of monitoring targets
	serviceNames := make([]string, 0)
	for _, target := range instance.MonitoringTargets.Targets {
		serviceNames = append(serviceNames, target.Service)
	}
	nameToID, err := d.monitoringTargetsEndpoints(serviceNames, instance.ComposePath())
	if err != nil {
		return err
	}
	// Remove monitoring targets
	for _, target := range instance.MonitoringTargets.Targets {
		endpoint, err := d.idToEndpoint(nameToID[target.Service], target.Path, target.Port)
		if err != nil {
			return err
		}
		networks, err := d.docker.ContainerNetworks(nameToID[target.Service])
		if err != nil {
			return err
		}
		if err = d.monitoringMgr.AddTarget(endpoint, instanceID, networks[0]); err != nil {
			return err
		}
	}

	return nil
}

func (d *WizDaemon) removeTarget(instanceID string) error {
	// Get monitoring targets
	instance, err := d.dataDir.Instance(instanceID)
	if err != nil {
		return err
	}
	// Get containerID of monitoring targets
	serviceNames := make([]string, 0)
	for _, target := range instance.MonitoringTargets.Targets {
		serviceNames = append(serviceNames, target.Service)
	}
	nameToID, err := d.monitoringTargetsEndpoints(serviceNames, instance.ComposePath())
	if err != nil {
		return err
	}
	// Remove monitoring targets
	for _, target := range instance.MonitoringTargets.Targets {
		endpoint, err := d.idToEndpoint(nameToID[target.Service], target.Path, target.Port)
		if err != nil {
			return err
		}
		networks, err := d.docker.ContainerNetworks(nameToID[target.Service])
		if err != nil {
			return err
		}
		if err = d.monitoringMgr.RemoveTarget(endpoint, networks[0]); err != nil {
			return err
		}
	}

	return nil
}
