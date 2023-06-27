package daemon

import "github.com/NethermindEth/egn/internal/compose"

// ComposeManager is an interface that defines methods for managing Docker Compose operations.
type ComposeManager interface {
	// Up starts the Docker Compose services defined in the Docker Compose file specified in the options.
	Up(opts compose.DockerComposeUpOptions) error

	// Stop stops the Docker Compose services defined in the Docker Compose file.
	Stop(opts compose.DockerComposeStopOptions) error

	// Down stops and removes the Docker Compose services defined in the Docker Compose file specified in the options.
	Down(opts compose.DockerComposeDownOptions) error
}
