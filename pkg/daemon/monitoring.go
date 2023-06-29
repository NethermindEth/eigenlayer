package daemon

import "github.com/NethermindEth/egn/internal/common"

type MonitoringManager interface {
	// InitStack initializes the monitoring stack.
	InitStack() error
	// AddTarget adds a new target to the monitoring stack.
	AddTarget(endpoint string) error
	// RemoveTarget removes a target from the monitoring stack.
	RemoveTarget(endpoint string) error
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
