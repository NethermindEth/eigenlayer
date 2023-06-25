package monitoring

import (
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/data"
	"github.com/NethermindEth/egn/internal/locker"
	"github.com/NethermindEth/egn/internal/monitoring/services/types"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

//go:embed script
var script embed.FS

var (
	ErrInitializingMonitoringMngr = errors.New("error initializing monitoring manager")
	ErrCheckingMonitoringStack    = errors.New("error checking monitoring stack status")
	ErrRunningMonitoringStack     = errors.New("error running monitoring stack")
)

// MonitoringManager manages the monitoring services. It provides methods for initializing the monitoring stack,
// adding and removing targets, running and stopping the monitoring stack, and checking the status of the monitoring stack.
type MonitoringManager struct {
	services       []ServiceAPI
	composeManager ComposeManager
	dockerManager  DockerManager
	stack          *data.MonitoringStack
}

// NewMonitoringManager creates a new MonitoringManager with the given services, compose manager, docker manager, file system, and locker.
func NewMonitoringManager(
	services []ServiceAPI,
	cmpMgr ComposeManager,
	dockerMgr DockerManager,
	fs afero.Fs,
	locker locker.Locker,
) *MonitoringManager {
	// Create stack
	datadir, err := data.NewDataDirDefault(fs, locker)
	if err != nil {
		log.Fatal(err)
	}
	stack, err := datadir.MonitoringStack()
	if err != nil {
		log.Fatal(err)
	}

	return &MonitoringManager{
		services:       services,
		composeManager: cmpMgr,
		dockerManager:  dockerMgr,
		stack:          stack,
	}
}

// InitStack initializes the monitoring stack by merging all environment variables, checking ports, setting up the stack and services, and creating containers.
func (m *MonitoringManager) InitStack() error {
	// Merge all dotEnv
	dotEnv := make(map[string]string)
	defaultPorts := make(map[string]uint16)
	for _, service := range m.services {
		for k, v := range service.DotEnv() {
			dotEnv[k] = v
			// Grab default ports
			if strings.HasSuffix(k, "_PORT") {
				// Cast string to uint16
				p, err := strconv.ParseUint(v, 10, 16)
				if err != nil {
					return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
				}
				defaultPorts[k] = uint16(p)
			}
		}
	}

	// Check ports
	ports, err := assignPorts("localhost", defaultPorts)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
	}
	for k, v := range ports {
		dotEnv[k] = strconv.Itoa(int(v))
	}

	// Intialize stack
	for _, service := range m.services {
		if err := service.Init(types.ServiceOptions{
			Stack:  m.stack,
			Dotenv: dotEnv,
		}); err != nil {
			return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
		}
	}
	if err = m.stack.Setup(dotEnv, script); err != nil {
		return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
	}

	// Setup services
	log.Info("Setting up monitoring stack...")
	for _, service := range m.services {
		if err = service.Setup(dotEnv); err != nil {
			return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
		}
	}

	// Create containers
	if err = m.composeManager.Create(compose.DockerComposeCreateOptions{Path: filepath.Join(m.stack.Path(), "docker-compose.yml")}); err != nil {
		return fmt.Errorf("%w: %w", ErrInitializingMonitoringMngr, err)
	}

	return nil
}

// AddTarget adds a new target to all services in the monitoring stack.
func (m *MonitoringManager) AddTarget(endpoint string) error {
	for _, service := range m.services {
		if err := service.AddTarget(endpoint); err != nil {
			return err
		}
	}
	return nil
}

// RemoveTarget removes a target from all services in the monitoring stack.
func (m *MonitoringManager) RemoveTarget(endpoint string) error {
	for _, service := range m.services {
		if err := service.RemoveTarget(endpoint); err != nil {
			return err
		}
	}
	return nil
}

// Run starts the monitoring stack by shutting down any existing stack and starting a new one.
func (m *MonitoringManager) Run() error {
	log.Info("Shutting down monitoring stack...")
	if err := m.composeManager.Down(compose.DockerComposeDownOptions{Path: filepath.Join(m.stack.Path(), "docker-compose.yml")}); err != nil {
		return fmt.Errorf("%w: %w", ErrRunningMonitoringStack, err)
	}

	log.Info("Starting monitoring stack...")
	if err := m.composeManager.Up(compose.DockerComposeUpOptions{Path: filepath.Join(m.stack.Path(), "docker-compose.yml")}); err != nil {
		return fmt.Errorf("%w: %w", ErrRunningMonitoringStack, err)
	}

	return nil
}

// Stop shuts down the monitoring stack.
func (m *MonitoringManager) Stop() error {
	log.Info("Shutting down monitoring stack...")
	if err := m.composeManager.Down(compose.DockerComposeDownOptions{Path: filepath.Join(m.stack.Path(), "docker-compose.yml")}); err != nil {
		return fmt.Errorf("%w: %w", ErrRunningMonitoringStack, err)
	}

	return nil
}

// Status checks the status of the containers in the monitoring stack and returns the status.
func (m *MonitoringManager) Status() (status common.Status, err error) {
	containers := []string{
		GrafanaContainerName,
		PrometheusContainerName,
		NodeExporterContainerName,
	}

	for _, container := range containers {
		status, err = m.dockerManager.ContainerStatus(container)
		if err != nil {
			return common.Unknown, fmt.Errorf("%w: %w", ErrCheckingMonitoringStack, err)
		}
		// running or restarting means the stack is running
		if status != common.Running && status != common.Restarting {
			return common.Broken,
				fmt.Errorf("%w: %s container is either paused, exited, or dead", ErrCheckingMonitoringStack, container)
		}
	}

	return status, nil
}

// InstallationStatus checks whether the monitoring stack is installed and returns the installation status.
func (m *MonitoringManager) InstallationStatus() (common.Status, error) {
	installed, err := m.stack.Installed()
	if err != nil {
		return common.Unknown, fmt.Errorf("%w: %w", ErrCheckingMonitoringStack, err)
	}
	if installed {
		return common.Installed, nil
	}

	return common.NotInstalled, nil
}
