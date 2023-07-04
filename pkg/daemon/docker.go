package daemon

type DockerManager interface {
	// ContainerIP returns the IP address of the container.
	ContainerIP(container string) (string, error)

	// ContainerNetworks returns the networks of a container.
	ContainerNetworks(container string) ([]string, error)
}
