package monitoring

import "github.com/NethermindEth/egn/internal/common"

// DockerManager is an interface for managing Docker containers.
type DockerManager interface {
	// ContainerStatus returns the status of a container.
	ContainerStatus(container string) (common.Status, error)
}
