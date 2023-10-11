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

	// LoadImageContext loads the given context.
	LoadImageContext(path string) (io.ReadCloser, error)

	// Run runs the given image with the given network and arguments.
	Run(image string, options docker.RunOptions) error

	// ContainerLogsMerged returns the merge of the logs of the given services.
	ContainerLogsMerged(ctx context.Context, w io.Writer, services map[string]string, opts docker.ContainerLogsMergedOptions) error

	// ImageRemove removes the given image.
	ImageRemove(image string) error

	// ImageExists checks if the given image exists.
	ImageExist(image string) (bool, error)
}
