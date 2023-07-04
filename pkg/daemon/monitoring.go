package daemon

import "github.com/NethermindEth/egn/internal/common"

type MonitoringManager interface {
	// Init initializes the monitoring stack. Assumes that the stack is already installed.
	Init() error

	// InstallStack installs the monitoring stack.
	InstallStack() error

	// AddTarget adds a new target to the monitoring stack.
	// The instanceID is used to identify the node in the monitoring stack.
	// The dockerNetwork is the name of the network the node is connected to.
	AddTarget(endpoint, instanceID, dockerNetwork string) error

	// RemoveTarget removes a target from the monitoring stack.
	// The dockerNetwork is the name of the network the node is connected to.
	RemoveTarget(endpoint, dockerNetwork string) error

	// Status returns the status of the monitoring stack.
	Status() (common.Status, error)

	// InstallationStatus returns the installation status of the monitoring stack.
	InstallationStatus() (common.Status, error)

	// Run runs the monitoring stack.
	Run() error

	// Stop stops the monitoring stack.
	Stop() error

	// Cleanup removes the monitoring stack. If force is true, it will remove the stack directly bypassing any checks.
	Cleanup(force bool) error
}
