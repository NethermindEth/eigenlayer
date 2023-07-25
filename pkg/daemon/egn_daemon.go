package daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/internal/compose"
	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/docker"
	"github.com/NethermindEth/eigenlayer/internal/locker"
	"github.com/NethermindEth/eigenlayer/internal/package_handler"
	"github.com/NethermindEth/eigenlayer/internal/utils"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/types"
	log "github.com/sirupsen/logrus"
)

// Checks that EgnDaemon implements Daemon.
var _ = Daemon(&EgnDaemon{})

// EgnDaemon is the main entrypoint for all the functionalities of the daemon.
type EgnDaemon struct {
	dataDir       *data.DataDir
	dockerCompose ComposeManager
	docker        DockerManager
	monitoringMgr MonitoringManager
	locker        locker.Locker
}

// NewDaemon create a new daemon instance.
func NewEgnDaemon(
	dataDir *data.DataDir,
	cmpMgr ComposeManager,
	dockerMgr DockerManager,
	mtrMgr MonitoringManager,
	locker locker.Locker,
) (*EgnDaemon, error) {
	return &EgnDaemon{
		dataDir:       dataDir,
		dockerCompose: cmpMgr,
		docker:        dockerMgr,
		monitoringMgr: mtrMgr,
		locker:        locker,
	}, nil
}

// Init initializes the Monitoring Stack. If install is true, it will install the Monitoring Stack if it is not installed.
// If run is true, it will run the Monitoring Stack if it is not running.
func (d *EgnDaemon) InitMonitoring(install, run bool) error {
	// Check if the monitoring stack is installed.
	installStatus, err := d.monitoringMgr.InstallationStatus()
	if err != nil {
		return err
	}
	log.Debugf("Monitoring stack installation status: %v", installStatus == common.Installed)
	// If the monitoring stack is not installed, install it.
	if installStatus == common.NotInstalled && install {
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
	}
	// Check if the monitoring stack is running.
	status, err := d.monitoringMgr.Status()
	if err != nil {
		log.Errorf("Monitoring stack status: unknown. Got error: %v", err)
	}
	// If the monitoring stack is not running, start it.
	if status != common.Running && status != common.Restarting && run {
		if err := d.monitoringMgr.Run(); err != nil {
			return err
		}
	} else if status != common.Running && status != common.Restarting && !run {
		// If the monitoring stack is not supposed to be running then exit.
		// This should change when the daemon runs as a real daemon.
		return nil
	}

	// Initialize monitoring stack if it is running.
	if err := d.monitoringMgr.Init(); err != nil {
		return err
	}

	// Add monitoring targets
	instanceIds, err := d.dataDir.ListInstances()
	if err != nil {
		return err
	}
	for _, instanceId := range instanceIds {
		if err := d.addTarget(instanceId); err != nil {
			return err
		}
	}

	return nil
}

// CleanMonitoring stops and uninstalls the Monitoring Stack
func (d *EgnDaemon) CleanMonitoring() error {
	// Check if the monitoring stack is installed.
	installStatus, err := d.monitoringMgr.InstallationStatus()
	if err != nil {
		return err
	}
	log.Debugf("Monitoring stack installation status: %v", installStatus == common.Installed)
	// If the monitoring stack is installed, uninstall it.
	if installStatus == common.Installed {
		if err := d.monitoringMgr.Cleanup(false); err != nil {
			return err
		}
	}
	return nil
}

// ListInstances implements Daemon.ListInstances.
func (d *EgnDaemon) ListInstances() ([]ListInstanceItem, error) {
	var result []ListInstanceItem
	instanceIds, err := d.dataDir.ListInstances()
	if err != nil {
		return result, err
	}
	for _, instanceId := range instanceIds {
		running, err := d.instanceRunning(instanceId)
		if err != nil {
			result = append(result, ListInstanceItem{
				ID:      instanceId,
				Health:  NodeHealthUnknown,
				Comment: fmt.Sprintf("Failed to get instance status: %v", err),
			})
			continue
		}
		var item ListInstanceItem
		if running {
			item = d.instanceHealth(instanceId)
		}
		item.ID = instanceId
		item.Running = running
		result = append(result, item)
	}
	return result, nil
}

func (d *EgnDaemon) instanceRunning(instanceId string) (bool, error) {
	instance, err := d.dataDir.Instance(instanceId)
	if err != nil {
		return false, err
	}
	composePath := instance.ComposePath()
	psOut, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:          composePath,
		Format:        "json",
		FilterRunning: true,
	})
	if err != nil {
		return false, err
	}
	var psServices []psServiceJSON
	if err = json.Unmarshal([]byte(psOut), &psServices); err != nil {
		return false, err
	}
	return len(psServices) > 0, nil
}

func (d *EgnDaemon) instanceHealth(instanceId string) (out ListInstanceItem) {
	out.ID = instanceId
	out.Health = NodeHealthUnknown // Default health is unknown

	// Get instance
	instance, err := d.dataDir.Instance(instanceId)
	if err != nil {
		return
	}

	if instance.APITarget == nil {
		// Instance does not have an API target
		out.Comment = "Instance's package does not specifies an API target for the AVS Specification Metrics's API"
		return
	}

	var psOut []composePsItem
	psOutRaw, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		ServiceName: instance.APITarget.Service,
		Path:        instance.ComposePath(),
		Format:      "json",
		All:         true,
	})
	if err != nil {
		out.Comment = fmt.Sprintf("Failed to get API container status: %v", err)
		return
	}
	err = json.Unmarshal([]byte(psOutRaw), &psOut)
	if err != nil {
		out.Comment = fmt.Sprintf("Failed to get API container status: %v", err)
		return
	}
	if len(psOut) == 0 {
		out.Comment = "No API container found"
		return
	}
	if psOut[0].State != "running" {
		out.Comment = "API container is " + psOut[0].State
		return
	}
	apiCtIP, err := d.docker.ContainerIP(psOut[0].Id)
	if err != nil {
		out.Comment = fmt.Sprintf("Failed to get API container IP: %v", err)
	}
	nodeHealth, err := checkHealth(apiCtIP, instance.APITarget.Port)
	if err != nil {
		out.Comment = fmt.Sprintf("API container is running but health check failed: %v", err)
	}
	out.Health = nodeHealth
	return
}

func checkHealth(ip string, port string) (NodeHealth, error) {
	url := fmt.Sprintf("http://%s:%s/eigen/node/health", ip, port)

	// HTTP client with timeout
	client := &http.Client{
		Timeout: time.Second * 10, // Timeout after 10 seconds
	}

	log.Debug("Checking health of node at ", url)
	resp, err := client.Get(url)
	if err != nil {
		return NodeHealthUnknown, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		return NodeHealthy, nil
	case 206:
		return NodePartiallyHealthy, nil
	case 503:
		return NodeUnhealthy, nil
	default:
		return NodeHealthUnknown, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

// Pull implements Daemon.Pull.
func (d *EgnDaemon) Pull(url string, version string, force bool) (result PullResult, err error) {
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
	// Set version
	if version == "" {
		version, err = pkgHandler.LatestVersion()
		if err != nil {
			return
		}
	}
	result.Version = version
	err = pkgHandler.CheckoutVersion(version)
	if err != nil {
		return
	}
	if err = pkgHandler.Check(); err != nil {
		return
	}
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
	result.HasPlugin, err = pkgHandler.HasPlugin()

	return result, err
}

// Install implements Daemon.Install.
func (d *EgnDaemon) Install(options InstallOptions) (string, error) {
	instanceId, tempDirID, err := d.install(options)
	return instanceId, d.postInstallation(instanceId, tempDirID, err)
}

func (d *EgnDaemon) LocalInstall(pkgTar io.Reader, options LocalInstallOptions) (string, error) {
	instanceId, tempDirID, err := d.localInstall(pkgTar, options)
	return instanceId, d.postInstallation(instanceId, tempDirID, err)
}

func (d *EgnDaemon) localInstall(pkgTar io.Reader, options LocalInstallOptions) (string, string, error) {
	instanceID := data.InstanceId(options.Name, options.Tag)
	// Check if instance already exists
	if d.dataDir.HasInstance(instanceID) {
		return instanceID, "", fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceID)
	}
	// Decompress package to temp folder
	tID := tempID(options.Name)
	tempPath, err := d.dataDir.InitTemp(tID)
	if err != nil {
		return instanceID, tID, err
	}
	err = utils.DecompressTarGz(pkgTar, tempPath)
	if err != nil {
		return instanceID, tID, err
	}
	// Init package handler from temp path
	pkgHandler := package_handler.NewPackageHandler(tempPath)
	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return instanceID, tID, err
	}
	// Select selectedProfile
	var selectedProfile *package_handler.Profile
	for _, pkgProfile := range pkgProfiles {
		if pkgProfile.Name == options.Profile {
			selectedProfile = &pkgProfile
			break
		}
	}
	if selectedProfile == nil {
		return instanceID, tID, fmt.Errorf("%w: %s", ErrProfileDoesNotExist, options.Profile)
	}
	// Validate profile
	err = selectedProfile.Validate()
	if err != nil {
		return instanceID, tID, err
	}
	// Build profile options
	profileOptions := make(map[string]Option, len(selectedProfile.Options))
	for _, o := range selectedProfile.Options {
		switch o.Type {
		case "str":
			profileOptions[o.Name] = NewOptionString(o)
		case "int":
			profileOptions[o.Name], err = NewOptionInt(o)
		case "float":
			profileOptions[o.Name], err = NewOptionFloat(o)
		case "bool":
			profileOptions[o.Name], err = NewOptionBool(o)
		case "path_dir":
			profileOptions[o.Name] = NewOptionPathDir(o)
		case "path_file":
			profileOptions[o.Name] = NewOptionPathFile(o)
		case "uri":
			profileOptions[o.Name] = NewOptionURI(o)
		case "select":
			profileOptions[o.Name] = NewOptionSelect(o)
		case "port":
			profileOptions[o.Name], err = NewOptionPort(o)
		case "id":
			profileOptions[o.Name] = NewOptionID(o)
		default:
			err = errors.New("unknown option type: " + o.Type)
			return instanceID, tID, err
		}
	}
	if err != nil {
		return instanceID, tID, err
	}

	// Build environment variables
	env := make(map[string]string, len(options.Options))
	for _, o := range profileOptions {
		if v, ok := options.Options[o.Name()]; ok {
			err := o.Set(v)
			if err != nil {
				return instanceID, tID, err
			}
			env[o.Target()] = o.Value()
		} else if o.Default() != "" {
			env[o.Target()] = o.Default()
		} else {
			return instanceID, tID, fmt.Errorf("%w: %s", ErrOptionWithoutValue, o.Name())
		}
	}
	// Get Monitoring targets
	monitoringTargets := make([]data.MonitoringTarget, 0)
	for _, target := range selectedProfile.Monitoring.Targets {
		if target.Port == nil {
			return instanceID, tID, ErrMonitoringTargetPortNotSet
		}
		mt := data.MonitoringTarget{
			Service: target.Service,
			Port:    strconv.Itoa(*target.Port),
			Path:    target.Path,
		}
		monitoringTargets = append(monitoringTargets, mt)
	}
	// Build plugin info
	var plugin *data.Plugin
	hasPlugin, err := pkgHandler.HasPlugin()
	if err != nil {
		return instanceID, tID, err
	}
	if hasPlugin {
		pkgPlugin, err := pkgHandler.Plugin()
		if err != nil {
			return instanceID, tID, err
		}
		plugin = &data.Plugin{
			Image:     pkgPlugin.Image,
			BuildFrom: pkgPlugin.BuildFrom,
		}
	}

	// Build API target info
	var apiTarget *data.APITarget
	if selectedProfile.API != nil {
		apiTarget = &data.APITarget{
			Service: selectedProfile.API.Service,
			Port:    strconv.Itoa(selectedProfile.API.Port),
		}
	}

	// Init instance
	instance := data.Instance{
		Name:              options.Name,
		Profile:           selectedProfile.Name,
		Version:           "v1.0.0",
		URL:               "http://localhost",
		Tag:               options.Tag,
		MonitoringTargets: data.MonitoringTargets{Targets: monitoringTargets},
		APITarget:         apiTarget,
		Plugin:            plugin,
	}
	if err = d.dataDir.InitInstance(&instance); err != nil {
		return instanceID, tID, err
	}
	if err = instance.Setup(env, pkgHandler.ProfilePath(instance.Profile)); err != nil {
		return instanceID, tID, err
	}

	// Create containers
	if err = d.dockerCompose.Create(compose.DockerComposeCreateOptions{
		Path:  instance.ComposePath(),
		Build: true,
	}); err != nil {
		return instanceID, tID, err
	}

	return instanceID, tID, nil
}

func (d *EgnDaemon) install(options InstallOptions) (string, string, error) {
	// Get temp folder ID
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

	// Build environment variables
	env := make(map[string]string, len(options.Options))
	for _, o := range options.Options {
		env[o.Target()] = o.Value()
	}
	// Get monitoring targets
	monitoringTargets := make([]data.MonitoringTarget, 0)
	for _, target := range selectedProfile.Monitoring.Targets {
		if target.Port == nil {
			return "", tID, ErrMonitoringTargetPortNotSet
		}
		mt := data.MonitoringTarget{
			Service: target.Service,
			Port:    strconv.Itoa(*target.Port),
			Path:    target.Path,
		}
		monitoringTargets = append(monitoringTargets, mt)
	}

	// Build plugin info
	var plugin *data.Plugin
	hasPlugin, err := pkgHandler.HasPlugin()
	if err != nil {
		return "", tID, err
	}
	if hasPlugin {
		pkgPlugin, err := pkgHandler.Plugin()
		if err != nil {
			return "", tID, err
		}
		plugin = &data.Plugin{
			Image:     pkgPlugin.Image,
			BuildFrom: pkgPlugin.BuildFrom,
		}
	}

	// Build API target info
	var apiTarget *data.APITarget
	if selectedProfile.API != nil {
		apiTarget = &data.APITarget{
			Service: selectedProfile.API.Service,
			Port:    strconv.Itoa(selectedProfile.API.Port),
		}
	}

	// Init instance
	instance := data.Instance{
		Name:              instanceName,
		Profile:           selectedProfile.Name,
		Version:           options.Version,
		URL:               options.URL,
		Tag:               options.Tag,
		MonitoringTargets: data.MonitoringTargets{Targets: monitoringTargets},
		APITarget:         apiTarget,
		Plugin:            plugin,
	}
	if err = d.dataDir.InitInstance(&instance); err != nil {
		return instanceId, tID, err
	}

	if err = instance.Setup(env, pkgHandler.ProfilePath(instance.Profile)); err != nil {
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

	return instanceId, tID, nil
}

func (d *EgnDaemon) postInstallation(instanceId string, tempDirID string, installErr error) error {
	if installErr != nil && instanceId != "" {
		// Cleanup if Install fails
		if cerr := d.uninstall(instanceId, false); cerr != nil {
			return fmt.Errorf("install failed: %w. Failed to cleanup after installation failure: %w", installErr, cerr)
		}
	}

	// Cleanup temp folder
	if rerr := d.dataDir.RemoveTemp(tempDirID); rerr != nil {
		if installErr != nil {
			return fmt.Errorf("install failed: %w. Failed to cleanup temporary folder after installation failure: %w", installErr, rerr)
		} else {
			return fmt.Errorf("install failed. Failed to cleanup temporary folder after installation failure: %w", rerr)
		}
	}
	return installErr
}

func (d *EgnDaemon) HasInstance(instanceID string) bool {
	return d.dataDir.HasInstance(instanceID)
}

// Run implements Daemon.Run.
func (d *EgnDaemon) Run(instanceID string) error {
	instancePath, err := d.dataDir.InstancePath(instanceID)
	if err != nil {
		return err
	}
	composePath := path.Join(instancePath, "docker-compose.yml")
	if err := d.dockerCompose.Up(compose.DockerComposeUpOptions{
		Path: composePath,
	}); err != nil {
		return err
	}

	return d.addTarget(instanceID)
}

// Stop implements Daemon.Stop.
func (d *EgnDaemon) Stop(instanceID string) error {
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
func (d *EgnDaemon) Uninstall(instanceID string) error {
	return d.uninstall(instanceID, true)
}

func (d *EgnDaemon) uninstall(instanceID string, down bool) error {
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

type composePsItem struct {
	Id    string `json:"ID"`
	Name  string `json:"Name"`
	State string `json:"State"`
}

// RunPlugin implements Daemon.RunPlugin.
func (d *EgnDaemon) RunPlugin(instanceId string, pluginArgs []string, options RunPluginOptions) error {
	instance, err := d.dataDir.Instance(instanceId)
	if err != nil {
		return err
	}
	composePath := instance.ComposePath()
	psOutputJSON, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:   composePath,
		Format: "json",
	})
	if err != nil {
		return err
	}
	var psOutput []composePsItem
	err = json.Unmarshal([]byte(psOutputJSON), &psOutput)
	if err != nil {
		return err
	}
	if len(psOutput) == 0 {
		return fmt.Errorf("%w: %s", ErrInstanceNotRunning, instanceId)
	}
	ct := psOutput[0]
	networks, err := d.docker.ContainerNetworks(ct.Id)
	if err != nil {
		return err
	}
	if len(networks) == 0 {
		return fmt.Errorf("%w: %s", ErrInstanceNotRunning, instanceId)
	}
	// Create plugin container
	var image string
	if instance.Plugin.Image != "" {
		// Pull image
		if err = d.docker.Pull(instance.Plugin.Image); err != nil {
			return err
		}
		image = instance.Plugin.Image
	} else if instance.Plugin.BuildFrom != "" {
		image = "eigen-plugin-" + instanceId
		// Build image
		if err = d.docker.BuildFromURI(instance.Plugin.BuildFrom, image); err != nil {
			return err
		}
	}
	network := networks[0]
	if options.HostNetwork {
		network = docker.NetworkHost
	}
	log.Infof("Running plugin with image %s on network %s", image, network)
	return d.docker.Run(image, network, pluginArgs)
}

// NodeLogs implements Daemon.NodeLogs.
func (d *EgnDaemon) NodeLogs(ctx context.Context, w io.Writer, instanceID string, opts NodeLogsOptions) error {
	i, err := d.dataDir.Instance(instanceID)
	if err != nil {
		return err
	}
	psRaw, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:   i.ComposePath(),
		Format: "json",
		All:    true,
	})
	if err != nil {
		return err
	}
	var ps []composePsItem
	err = json.Unmarshal([]byte(psRaw), &ps)
	if err != nil {
		return err
	}
	services := make(map[string]string, len(ps))
	for _, p := range ps {
		services[p.Name] = p.Id
	}

	return d.docker.ContainerLogsMerged(ctx, w, services, docker.ContainerLogsMergedOptions{
		Follow:     opts.Follow,
		Since:      opts.Since,
		Until:      opts.Until,
		Timestamps: opts.Timestamps,
		Tail:       opts.Tail,
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

type psServiceJSON struct {
	ID      string `json:"ID"`
	Service string `json:"Service"`
}

func (d *EgnDaemon) monitoringTargetsEndpoints(serviceNames []string, composePath string) (map[string]string, error) {
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
		return nil, fmt.Errorf("it seems the output ")
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

	// Validate that all monitoring targets were found
	for _, serviceName := range serviceNames {
		if _, ok := monitoringTargets[serviceName]; !ok {
			return nil, fmt.Errorf("monitoring target %s not found, there is not such service running in the docker compose stack", serviceName)
		}
	}

	return monitoringTargets, nil
}

func (d *EgnDaemon) idToIP(id string) (string, error) {
	ip, err := d.docker.ContainerIP(id)
	if err != nil {
		return "", err
	}
	return ip, nil
}

func (d *EgnDaemon) addTarget(instanceID string) error {
	// Check if the monitoring stack is installed.
	installStatus, err := d.monitoringMgr.InstallationStatus()
	if err != nil {
		return err
	}
	if installStatus != common.Installed {
		return nil
	}
	// Check if the monitoring stack is running.
	status, err := d.monitoringMgr.Status()
	if err != nil {
		return fmt.Errorf("monitoring stack status: unknown. Got error: %v", err)
	}
	// If the monitoring stack is not running, skip.
	if status != common.Running && status != common.Restarting {
		return nil
	}

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
		endpoint, err := d.idToIP(nameToID[target.Service])
		if err != nil {
			return err
		}
		if endpoint == "" {
			// This means the container is not running. Skip.
			continue
		}
		networks, err := d.docker.ContainerNetworks(nameToID[target.Service])
		if err != nil {
			return err
		}
		port, err := strconv.ParseUint(target.Port, 10, 16)
		if err != nil {
			return err
		}

		if err = d.monitoringMgr.AddTarget(types.MonitoringTarget{
			Host: endpoint,
			Port: uint16(port),
			Path: target.Path,
		}, instanceID, networks[0]); err != nil {
			return err
		}
	}

	return nil
}

func (d *EgnDaemon) removeTarget(instanceID string) error {
	// Check if the monitoring stack is installed.
	installStatus, err := d.monitoringMgr.InstallationStatus()
	if err != nil {
		return err
	}
	if installStatus != common.Installed {
		return nil
	}
	// Check if the monitoring stack is running.
	status, err := d.monitoringMgr.Status()
	if err != nil {
		return fmt.Errorf("monitoring stack status: unknown. Got error: %v", err)
	}
	// If the monitoring stack is not running, skip.
	if status != common.Running && status != common.Restarting {
		return nil
	}

	// Remove target from monitoring stack
	return d.monitoringMgr.RemoveTarget(instanceID)
}
