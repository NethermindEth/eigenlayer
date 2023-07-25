package daemon

import (
	"context"
	"io"

	"github.com/NethermindEth/eigenlayer/internal/docker"
)

type DockerManager interface {
	// ContainerIP returns the IP address of the container.
	ContainerIP(container string) (string, error)

	// ContainerNetworks returns the networks of a container.
	ContainerNetworks(container string) ([]string, error)

	// Pull pulls the given image.
	Pull(image string) error

	// Build builds the given image from the given remote and sets the given tag.
	BuildFromURI(remote string, tag string) error

	// Run runs the given image with the given network and arguments.
	Run(image string, network string, args []string, mounts []docker.Mount) error

	// ContainerLogsMerged returns the merge of the logs of the given services.
	ContainerLogsMerged(ctx context.Context, w io.Writer, services map[string]string, opts docker.ContainerLogsMergedOptions) error
}
