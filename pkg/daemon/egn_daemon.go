package daemon

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/maps"
	"golang.org/x/mod/semver"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/internal/compose"
	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/docker"
	hardwarechecker "github.com/NethermindEth/eigenlayer/internal/hardware_checker"
	"github.com/NethermindEth/eigenlayer/internal/locker"
	"github.com/NethermindEth/eigenlayer/internal/package_handler"
	"github.com/NethermindEth/eigenlayer/internal/profile"
	"github.com/NethermindEth/eigenlayer/internal/utils"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/types"
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
		log.Debugf("Monitoring stack status: unknown. Got error: %v", err)
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
	instances, err := d.dataDir.ListInstances()
	if err != nil {
		return err
	}
	for _, instance := range instances {
		if err := d.addTarget(instance.ID()); err != nil {
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
	instances, err := d.dataDir.ListInstances()
	if err != nil {
		return result, err
	}
	for _, instance := range instances {
		running, err := d.instanceRunning(instance.ID())
		if err != nil {
			result = append(result, ListInstanceItem{
				ID:      instance.ID(),
				Health:  NodeHealthUnknown,
				Comment: fmt.Sprintf("Failed to get instance status: %v", err),
				Version: instance.Version,
				Commit:  instance.Commit,
			})
			continue
		}
		var item ListInstanceItem
		if running {
			item = d.instanceHealth(instance.ID())
		}
		item.ID = instance.ID()
		item.Running = running
		item.Version = instance.Version
		item.Commit = instance.Commit
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
	psServices, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:          composePath,
		Format:        "json",
		FilterRunning: true,
	})
	if err != nil {
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

	psServices, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		ServiceName: instance.APITarget.Service,
		Path:        instance.ComposePath(),
		Format:      "json",
		All:         true,
	})
	if err != nil {
		out.Comment = fmt.Sprintf("Failed to get API container status: %v", err)
		return
	}
	if len(psServices) == 0 {
		out.Comment = "No API container found"
		return
	}
	if psServices[0].State != "running" {
		out.Comment = "API container is " + psServices[0].State
		return
	}
	apiCtIP, err := d.docker.ContainerIP(psServices[0].Id)
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
func (d *EgnDaemon) Pull(url string, ref PullTarget, force bool) (result PullResult, err error) {
	pkgHandler, err := d.pullPackage(url, force)
	if err != nil {
		return
	}
	if ref.Version != "" {
		result.Version = ref.Version
		// Set version
		err = pkgHandler.CheckoutVersion(ref.Version)
		if err != nil {
			return
		}
	} else if ref.Commit != "" {
		err = pkgHandler.CheckoutCommit(ref.Commit)
		if err != nil {
			return
		}
	} else {
		var latestVersion string
		latestVersion, err = pkgHandler.LatestVersion()
		if err != nil {
			return
		}
		result.Version = latestVersion
		err = pkgHandler.CheckoutVersion(latestVersion)
		if err != nil {
			return
		}
	}
	// Get AVS name
	result.Name, err = pkgHandler.Name()
	if err != nil {
		return
	}
	// Get Spec version
	result.SpecVersion, err = pkgHandler.SpecVersion()
	if err != nil {
		return
	}
	// Get commit hash
	result.Commit, err = pkgHandler.CurrentCommitHash()
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
		options, err := optionsFromProfile(&profile)
		if err != nil {
			return PullResult{}, err
		}
		profileOptions[profile.Name] = options
	}
	result.Options = profileOptions
	result.HasPlugin, err = pkgHandler.HasPlugin()

	requirements := make(map[string]HardwareRequirements, len(profiles))
	for _, profile := range profiles {
		req, err := pkgHandler.HardwareRequirements(profile.Name)
		if err != nil {
			continue
		}
		requirements[profile.Name] = HardwareRequirements{
			MinCPUCores:                 req.MinCPUCores,
			MinRAM:                      req.MinRAM,
			MinFreeSpace:                req.MinFreeSpace,
			StopIfRequirementsAreNotMet: req.StopIfRequirementsAreNotMet,
		}
	}
	result.HardwareRequirements = requirements

	return result, err
}

func (d *EgnDaemon) PullUpdate(instanceID string, ref PullTarget) (PullUpdateResult, error) {
	if !d.dataDir.HasInstance(instanceID) {
		return PullUpdateResult{}, fmt.Errorf("%w: %s", ErrInstanceNotFound, instanceID)
	}
	instance, err := d.dataDir.Instance(instanceID)
	if err != nil {
		return PullUpdateResult{}, err
	}
	pkgHandler, err := d.pullPackage(instance.URL, true)
	if err != nil {
		return PullUpdateResult{}, err
	}
	if ref.Version != "" {
		// Check if the new version is greater than the current version
		versionCompare := semver.Compare(ref.Version, instance.Version)
		if versionCompare == 0 {
			return PullUpdateResult{}, fmt.Errorf("%w: %s", ErrVersionAlreadyInstalled, ref.Version)
		}
		if versionCompare != 1 {
			return PullUpdateResult{}, fmt.Errorf("%w: %s, must be grater than the current version", ErrInvalidUpdateVersion, ref.Version)
		}
		err = pkgHandler.CheckoutVersion(ref.Version)
		if err != nil {
			return PullUpdateResult{}, err
		}
	} else if ref.Commit != "" {
		err := pkgHandler.CheckoutCommit(ref.Commit)
		if err != nil {
			return PullUpdateResult{}, err
		}
	} else {
		latestVersion, err := pkgHandler.LatestVersion()
		if err != nil {
			return PullUpdateResult{}, err
		}
		versionCompare := semver.Compare(latestVersion, instance.Version)
		if versionCompare == 0 {
			return PullUpdateResult{}, fmt.Errorf("%w: %s", ErrVersionAlreadyInstalled, latestVersion)
		}
		if versionCompare != 1 {
			return PullUpdateResult{}, fmt.Errorf("%w: %s, must be grater than the current version", ErrInvalidUpdateVersion, ref.Version)
		}
		err = pkgHandler.CheckoutVersion(latestVersion)
		if err != nil {
			return PullUpdateResult{}, err
		}
	}
	// Get new commit hash
	newCommit, err := pkgHandler.CurrentCommitHash()
	if err != nil {
		return PullUpdateResult{}, err
	}

	if instance.Commit == newCommit {
		return PullUpdateResult{}, fmt.Errorf("%w: %s", ErrVersionAlreadyInstalled, newCommit)
	}

	// Check commit precedence
	ok, err := pkgHandler.CommitPrecedence(instance.Commit, newCommit)
	if err != nil {
		return PullUpdateResult{}, err
	}
	if !ok {
		return PullUpdateResult{}, fmt.Errorf("%w: current commit %s is not previous to the update commit %s", ErrInvalidUpdateCommit, instance.Commit, ref.Commit)
	}
	// Get new version
	newVersion, err := pkgHandler.CurrentVersion()
	if err != nil {
		return PullUpdateResult{}, err
	}

	// Get new options
	profileNew, err := pkgHandler.Profile(instance.Profile)
	if err != nil {
		return PullUpdateResult{}, err
	}
	optionsNew, err := optionsFromProfile(profileNew)
	if err != nil {
		return PullUpdateResult{}, err
	}
	// Get old options with its values
	profileOld, err := instance.ProfileFile()
	if err != nil {
		return PullUpdateResult{}, err
	}
	optionsOld, err := optionsFromProfile(profileOld)
	if err != nil {
		return PullUpdateResult{}, err
	}
	valuesOld, err := instance.Env()
	if err != nil {
		return PullUpdateResult{}, err
	}
	for _, o := range optionsOld {
		if v, ok := valuesOld[o.Target()]; ok {
			err := o.Set(v)
			if err != nil {
				return PullUpdateResult{}, fmt.Errorf("error setting old option value: %v", err)
			}
		} else {
			return PullUpdateResult{}, fmt.Errorf("%w: old option %s", ErrOptionWithoutValue, o.Name())
		}
	}

	mergedOptions, err := mergeOptions(optionsOld, optionsNew)
	if err != nil {
		return PullUpdateResult{}, err
	}

	return PullUpdateResult{
		Name:          instance.Name,
		Tag:           instance.Tag,
		Url:           instance.URL,
		Profile:       instance.Profile,
		HasPlugin:     instance.Plugin != nil,
		OldVersion:    instance.Version,
		NewVersion:    newVersion,
		OldCommit:     instance.Commit,
		NewCommit:     newCommit,
		OldOptions:    optionsOld,
		NewOptions:    optionsNew,
		MergedOptions: mergedOptions,
	}, nil
}

// mergeOptions merges the old options with the new ones following the next rules:
//
//  1. New option is not present in the old options: New option is added to the
//     merged options without a value.
//
// 2. New option is present in the old options:
//
//	2.1: Old option value is valid for the new option: the new option is added
//	     to the merged options using the old option value.
//
//	2.2: Old option value is not valid for the new option: the new option is
//	     added to the merged options without a value and probably the user
//	     will need to fill it again or use its default value.
func mergeOptions(oldOptions, newOptions []Option) ([]Option, error) {
	// mergedOptions will contain the result of merging the new options with the
	// old ones
	var mergedOptions []Option
	for _, oNew := range newOptions {
		var oOld Option
		for _, o := range oldOptions {
			// XXX: Do we need to check the option equality with the Name or the Target?
			if o.Name() == oNew.Name() {
				oOld = o
				break
			}
		}
		if oOld == nil {
			// Option does not exist previously
			mergedOptions = append(mergedOptions, oNew)
		} else {
			// Option exists previously. Try to set the old value
			oldValue, err := oOld.Value()
			if err != nil {
				// Old option is expected to have a value but it does not.
				return mergedOptions, err
			}
			err = oNew.Set(oldValue)
			if err != nil {
				// Old value is not valid for same option in the new version. This
				// option should be filled by the user again.
				log.Debugf("Option %s value %s is not valid for the new version. error: %s", oOld.Name(), oldValue, err.Error())
			} else {
				// Old value is valid for the new version and we can use it.
				log.Debugf("Option %s value %s is valid for the new version", oOld.Name(), oldValue)
			}
			mergedOptions = append(mergedOptions, oNew)
		}
	}
	return mergedOptions, nil
}

func (d *EgnDaemon) pullPackage(url string, force bool) (*package_handler.PackageHandler, error) {
	tID := tempID(url)
	if force {
		err := d.dataDir.RemoveTemp(tID)
		if err != nil {
			return nil, err
		}
	}
	tempPath, err := d.dataDir.InitTemp(tID)
	if err != nil {
		return nil, err
	}
	return package_handler.NewPackageHandlerFromURL(package_handler.NewPackageHandlerOptions{
		Path: tempPath,
		URL:  url,
	})
}

// Install implements Daemon.Install.
func (d *EgnDaemon) Install(options InstallOptions) (string, error) {
	instanceId, tempDirID, err := d.remoteInstall(options)
	return instanceId, d.postInstallation(instanceId, tempDirID, err)
}

func (d *EgnDaemon) LocalInstall(pkgTar io.Reader, options LocalInstallOptions) (string, error) {
	instanceId, tempDirID, err := d.localInstall(pkgTar, options)
	return instanceId, d.postInstallation(instanceId, tempDirID, err)
}

func (d *EgnDaemon) localInstall(pkgTar io.Reader, options LocalInstallOptions) (string, string, error) {
	// Decompress package to temp folder
	tID := tempID(options.Name)
	tempPath, err := d.dataDir.InitTemp(tID)
	if err != nil {
		return "", tID, err
	}
	err = utils.DecompressTarGz(pkgTar, tempPath)
	if err != nil {
		return "", tID, err
	}

	// Init package handler from temp path
	pkgHandler := package_handler.NewPackageHandler(tempPath)

	// Get Name
	name, err := pkgHandler.Name()
	if err != nil {
		return "", tID, err
	}

	// Get Instance ID
	instanceID := data.InstanceId(name, options.Tag)
	// Check if instance already exists
	if d.dataDir.HasInstance(instanceID) {
		return instanceID, "", fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceID)
	}

	// Get Spec version
	specVersion, err := pkgHandler.SpecVersion()
	if err != nil {
		return instanceID, tID, err
	}
	// Get profiles
	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return instanceID, tID, err
	}
	// Select selectedProfile
	var selectedProfile *profile.Profile
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
		default:
			err = errors.New("unknown option type: " + o.Type)
			return instanceID, tID, err
		}
	}
	if err != nil {
		return instanceID, tID, err
	}

	// Build environment variables
	env, err := pkgHandler.DotEnv(selectedProfile.Name)
	if err != nil {
		return instanceID, tID, err
	}
	optionsEnv := make(map[string]string, len(options.Options))
	for _, o := range profileOptions {
		if v, ok := options.Options[o.Name()]; ok {
			err := o.Set(v)
			if err != nil {
				return instanceID, tID, err
			}
			optionsEnv[o.Target()] = v
		} else if o.Default() != "" {
			optionsEnv[o.Target()] = o.Default()
		} else {
			return instanceID, tID, fmt.Errorf("%w: %s", ErrOptionWithoutValue, o.Name())
		}
	}
	maps.Copy(env, optionsEnv)

	installOptions := InstallOptions{
		Profile:     options.Profile,
		Tag:         options.Tag,
		URL:         "http://localhost",
		Version:     "local",
		SpecVersion: specVersion,
		Commit:      "local",
	}
	return d.install(options.Name, instanceID, tID, pkgHandler, selectedProfile, env, installOptions)
}

func (d *EgnDaemon) remoteInstall(options InstallOptions) (string, string, error) {
	// Get temp folder ID
	tID := tempID(options.URL)
	tempPath, err := d.dataDir.TempPath(tID)
	if err != nil {
		return "", tID, err
	}

	instanceID := data.InstanceId(options.Name, options.Tag)

	if d.dataDir.HasInstance(instanceID) {
		return instanceID, tID, fmt.Errorf("%w: %s", ErrInstanceAlreadyExists, instanceID)
	}

	// Init package handler from temp path
	pkgHandler := package_handler.NewPackageHandler(tempPath)
	if options.Version != "" {
		// Check if selected version is valid
		if err := pkgHandler.HasVersion(options.Version); err != nil {
			return instanceID, tID, err
		}
		if err = pkgHandler.CheckoutVersion(options.Version); err != nil {
			return instanceID, tID, err
		}
	} else if options.Commit != "" {
		err := pkgHandler.CheckoutCommit(options.Commit)
		if err != nil {
			return instanceID, tID, err
		}
	} else {
		return instanceID, tID, fmt.Errorf("%w: %s", ErrVersionOrCommitNotSet, options.URL)
	}

	pkgProfiles, err := pkgHandler.Profiles()
	if err != nil {
		return instanceID, tID, err
	}
	var selectedProfile *profile.Profile
	// Check if selected profile is valid
	for _, pkgProfile := range pkgProfiles {
		if pkgProfile.Name == options.Profile {
			selectedProfile = &pkgProfile
			break
		}
	}
	if selectedProfile == nil {
		return instanceID, tID, fmt.Errorf("%w: %s", ErrProfileDoesNotExist, options.Profile)
	}

	// Build environment variables
	env, err := pkgHandler.DotEnv(selectedProfile.Name)
	if err != nil {
		return instanceID, tID, err
	}
	optionsEnv := make(map[string]string, len(options.Options))
	for _, o := range options.Options {
		oValue, err := o.Value()
		if err != nil {
			return instanceID, tID, err
		}
		env[o.Target()] = oValue
	}
	maps.Copy(env, optionsEnv)

	return d.install(options.Name, instanceID, tID, pkgHandler, selectedProfile, env, options)
}

func (d *EgnDaemon) install(
	instanceName, instanceID, tID string,
	pkgHandler *package_handler.PackageHandler,
	selectedProfile *profile.Profile,
	env map[string]string,
	options InstallOptions,
) (string, string, error) {
	err := pkgHandler.CheckComposeProject(selectedProfile.Name, env)
	if err != nil {
		return instanceID, tID, err
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
	plugin, err := d.getPluginData(d.dataDir, pkgHandler, instanceID)
	if err != nil {
		return instanceID, tID, err
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
		SpecVersion:       options.SpecVersion,
		Commit:            options.Commit,
		URL:               options.URL,
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
	// TODO: Log Create output and log to wait as containers might be built
	if err = d.dockerCompose.Create(compose.DockerComposeCreateOptions{
		Path:  instance.ComposePath(),
		Build: true,
	}); err != nil {
		return instanceID, tID, err
	}

	return instanceID, tID, nil
}

func (d *EgnDaemon) getPluginData(dataDir *data.DataDir, pkgHandler *package_handler.PackageHandler, instanceID string) (*data.Plugin, error) {
	hasPlugin, err := pkgHandler.HasPlugin()
	if err != nil {
		return nil, err
	}
	if !hasPlugin {
		return nil, nil
	}
	pkgPlugin, err := pkgHandler.Plugin()
	if err != nil {
		return nil, err
	}
	return &data.Plugin{
		Image: pkgPlugin.Image,
	}, nil
}

func (d *EgnDaemon) postInstallation(instanceId string, tempDirID string, installErr error) error {
	if installErr != nil && !errors.Is(installErr, ErrInstanceAlreadyExists) {
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
	instancePath, err := d.dataDir.InstancePath(instanceID)
	if err != nil {
		if errors.Is(err, data.ErrInstanceNotFound) {
			log.Warnf("Instance %s not found. It may be due to a incomplete instance installation process.", instanceID)
			return nil
		}
		return err
	}

	if err := d.dataDir.RemovePluginContext(instanceID); err != nil {
		return err
	}

	if err := d.removeTarget(instanceID); err != nil {
		if errors.Is(err, monitoring.ErrNonexistingTarget) {
			log.Warnf("Monitoring target for instance %s not found. It may be due to an incomplete instance installation process or because the instance was never started.", instanceID)
		} else {
			return err
		}
	}

	if down {
		composePath := path.Join(instancePath, "docker-compose.yml")
		// docker compose down
		if err = d.dockerCompose.Down(compose.DockerComposeDownOptions{
			Path:    composePath,
			Volumes: true,
		}); err != nil {
			return err
		}
	}

	// remove instance directory
	return d.dataDir.RemoveInstance(instanceID)
}

// CheckHardwareRequirements implements Daemon.CheckHardwareRequirements
func (d *EgnDaemon) CheckHardwareRequirements(req HardwareRequirements) (bool, error) {
	metrics, err := hardwarechecker.GetMetrics()
	if err != nil {
		return false, err
	}
	requirements := hardwarechecker.HardwareMetrics{
		CPU:       float64(req.MinCPUCores),
		RAM:       float64(req.MinRAM),
		DiskSpace: float64(req.MinFreeSpace),
	}
	return metrics.Meets(requirements), nil
}

// RunPlugin implements Daemon.RunPlugin.
func (d *EgnDaemon) RunPlugin(instanceId string, pluginArgs []string, options RunPluginOptions) error {
	instance, err := d.dataDir.Instance(instanceId)
	if err != nil {
		return err
	}
	if instance.Plugin == nil {
		return fmt.Errorf("%w: %s", ErrInstanceHasNoPlugin, instanceId)
	}
	network := docker.NetworkHost
	if !options.HostNetwork {
		composePath := instance.ComposePath()
		psServices, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
			FilterRunning: true,
			Path:          composePath,
			Format:        "json",
		})
		if err != nil {
			return err
		}
		if len(psServices) == 0 {
			return fmt.Errorf("%w: %s", ErrInstanceNotRunning, instanceId)
		}
		ct := psServices[0]
		networks, err := d.docker.ContainerNetworks(ct.Id)
		if err != nil {
			return err
		}
		if len(networks) == 0 {
			return fmt.Errorf("%w: %s", ErrInstanceNotRunning, instanceId)
		}
		network = networks[0]
	}
	// XXX: Pull is removed to support local images that are already pulled
	// err = d.docker.Pull(instance.Plugin.Image)
	// if err != nil {
	// 	return err
	// }
	if !options.NoDestroyImage {
		defer func() {
			if err := d.docker.ImageRemove(instance.Plugin.Image); err != nil {
				log.Errorf("Failed to destroy plugin image %s: %v", instance.Plugin.Image, err)
			}
		}()
	}
	log.Infof("Running plugin with image %s on network %s", instance.Plugin.Image, network)
	mounts := make([]docker.Mount, 0, len(options.Binds)+len(options.Volumes))
	for src, dst := range options.Binds {
		_, err := os.Stat(src)
		if os.IsNotExist(err) {
			if filepath.Ext(src) != "" {
				return fmt.Errorf("bound file %s does not exist", src)
			}
			if err := os.MkdirAll(src, 0o755); err != nil {
				return fmt.Errorf("failed to create bound directory %s: %v", src, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to stat bound source %s: %v", src, err)
		}
		mounts = append(mounts, docker.Mount{
			Type:   docker.VolumeTypeBind,
			Source: src,
			Target: dst,
		})
	}
	for src, dst := range options.Volumes {
		mounts = append(mounts, docker.Mount{
			Type:   docker.VolumeTypeVolume,
			Source: src,
			Target: dst,
		})
	}
	return d.docker.Run(instance.Plugin.Image, network, pluginArgs, mounts)
}

// NodeLogs implements Daemon.NodeLogs.
func (d *EgnDaemon) NodeLogs(ctx context.Context, w io.Writer, instanceID string, opts NodeLogsOptions) error {
	i, err := d.dataDir.Instance(instanceID)
	if err != nil {
		return err
	}
	psServices, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:   i.ComposePath(),
		Format: "json",
		All:    true,
	})
	if err != nil {
		return err
	}
	services := make(map[string]string, len(psServices))
	for _, p := range psServices {
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

func tempID(url string) string {
	tempHash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(tempHash[:])
}

func (d *EgnDaemon) monitoringTargetsEndpoints(serviceNames []string, composePath string) (map[string]string, error) {
	psServices, err := d.dockerCompose.PS(compose.DockerComposePsOptions{
		Path:   composePath,
		Format: "json",
		All:    true,
	})
	if err != nil {
		return nil, err
	}

	// Get containerID of monitoring targets
	monitoringTargets := make(map[string]string)
	for _, serviceName := range serviceNames {
		for _, psService := range psServices {
			if psService.Service == serviceName {
				monitoringTargets[serviceName] = psService.Id
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

// addTarget adds an instance to the monitoring stack
// If the monitoring stack is not installed or running, it does nothing
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

		labels := map[string]string{
			monitoring.InstanceIDLabel:  instanceID,
			monitoring.CommitHashLabel:  instance.Commit,
			monitoring.AVSNameLabel:     instance.Name,
			monitoring.AVSVersionLabel:  instance.Version,
			monitoring.SpecVersionLabel: instance.SpecVersion,
		}
		if err = d.monitoringMgr.AddTarget(types.MonitoringTarget{
			Host: endpoint,
			Port: uint16(port),
			Path: target.Path,
		}, labels, networks[0]); err != nil {
			return err
		}
	}

	return nil
}

// removeTarget removes the instance from the monitoring stack.
// If the monitoring stack is not installed or not running, it does nothing.
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
