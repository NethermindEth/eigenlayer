package types

import "github.com/NethermindEth/eigenlayer/internal/data"

// ServiceOptions defines the options for initializing a monitoring service. It includes a reference to the monitoring stack
// and a map of environment variables.
type ServiceOptions struct {
	// Stack is a reference to the monitoring stack that the service is a part of.
	Stack *data.MonitoringStack

	// Dotenv is a map of environment variables for the service. The keys are the variable names and the values are the variable values.
	Dotenv map[string]string
}
