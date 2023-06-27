package monitoring

import "errors"

var (
	ErrInitializingMonitoringMngr = errors.New("error initializing monitoring manager")
	ErrCheckingMonitoringStack    = errors.New("error checking monitoring stack status")
	ErrRunningMonitoringStack     = errors.New("error running monitoring stack")
)
