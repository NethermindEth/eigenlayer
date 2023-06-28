package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/docker/docker/api/types"
	dockerCt "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/utils"
)

// NewDockerManager returns a new instance of DockerManager
func NewDockerManager(dockerClient client.APIClient) *DockerManager {
	return &DockerManager{dockerClient}
}

// DockerManager is an interface to the Docker Daemon for managing docker containers
type DockerManager struct {
	dockerClient client.APIClient
}

// Image retrieves the image name associated with a specified Docker container.
// The function accepts a string parameter 'container', which represents the name or ID of the Docker container.
// It uses the ContainerInspect method of the Docker client to fetch the container's information.
// If the container is found and no error occurs, the function returns the image name as a string.
// If the container is not found or an error occurs during the inspection, the function returns an empty string and the error.
func (d *DockerManager) Image(container string) (string, error) {
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), container)
	if err != nil {
		return "", err
	}
	return ctInfo.Image, nil
}

// Start initiates the start process of a specified Docker container.
// The function accepts a string parameter 'container', which represents the name or ID of the Docker container.
// It uses the ContainerStart method of the Docker client to start the container.
// If the container is successfully started, the function returns nil.
// If an error occurs during the start process, the function wraps the error with a custom message and returns it.
func (d *DockerManager) Start(container string) error {
	if err := d.dockerClient.ContainerStart(context.Background(), container, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("%w %s: %s", ErrStartingContainer, container, err)
	}
	return nil
}

// Stop attempts to stop a specified Docker container.
// The function first inspects the container to check if it's running. If the container is not found, it returns nil.
// If the container is running, it attempts to stop the container and returns any error that occurs during the process.
func (d *DockerManager) Stop(container string) error {
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), container)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return err
	}
	if ctInfo.State.Running || ctInfo.State.Restarting {
		log.Infof("Stopping service: %s, currently on %s status", container, ctInfo.State.Status)
		timeout := 5 * int(time.Minute)
		if err := d.dockerClient.ContainerStop(context.Background(), ctInfo.ID, dockerCt.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			return fmt.Errorf("%w %s: %s", ErrStoppingContainer, container, err)
		}
	}
	return nil
}

// ContainerID retrieves the ID of a specified Docker container name.
// The function lists all containers and filters them by name. If a container with the specified name is found, its ID is returned.
// If no container with the specified name is found, the function returns an error.
func (d *DockerManager) ContainerID(containerName string) (string, error) {
	containers, err := d.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", containerName)),
	})
	if err != nil {
		return "", err
	}
	for _, c := range containers {
		if utils.Contains(c.Names, "/"+containerName) {
			return c.ID, nil
		}
	}
	return "", fmt.Errorf("%w: %s", ErrContainerNotFound, containerName)
}

// Pull pulls a specified Docker image.
// The function attempts to pull the image and returns any error that occurs during the process.
func (d *DockerManager) Pull(image string) error {
	log.Debugf("Pulling image: %s", image)
	_, err := d.dockerClient.ImagePull(context.Background(), image, types.ImagePullOptions{})
	return err
}

// ContainerLogs retrieves the logs of a specified Docker container.
// The function accepts a string parameter 'container', which represents the name or ID of the Docker container.
// The function attempts to fetch the logs and returns them as a string. If an error occurs during the process, it returns the error.
func (d *DockerManager) ContainerLogs(container string) (string, error) {
	logReader, err := d.dockerClient.ContainerLogs(context.Background(), container, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return "", err
	}
	defer logReader.Close()

	logs, err := io.ReadAll(logReader)
	if err == nil {
		log.Debugf("Container logs: %s", string(logs))
	}
	return string(logs), err
}

// Wait waits for a specified Docker container to reach a certain condition.
// The function returns two channels: one for the wait response and one for any error that occurs during the wait process.
func (d *DockerManager) Wait(container string, condition WaitCondition) (<-chan WaitResponse, <-chan error) {
	dwrChan, err := d.dockerClient.ContainerWait(context.Background(), container, dockerCt.WaitCondition(condition))

	wrChan := make(chan WaitResponse)
	go func() {
		for r := range dwrChan {
			var waitExitError *WaitExitError
			if r.Error != nil {
				waitExitError = &WaitExitError{Message: r.Error.Message}
			}
			wrChan <- WaitResponse{
				StatusCode: r.StatusCode,
				Error:      waitExitError,
			}
		}
		close(wrChan)
	}()
	return wrChan, err
}

// ContainerStatus retrieves the status of a specified Docker container.
func (d *DockerManager) ContainerStatus(container string) (common.Status, error) {
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), container)
	if err != nil {
		return common.Unknown, err
	}
	switch ctInfo.State.Status {
	case "created":
		return common.Created, nil
	case "running":
		return common.Running, nil
	case "paused":
		return common.Paused, nil
	case "restarting":
		return common.Restarting, nil
	case "removing":
		return common.Removing, nil
	case "exited":
		return common.Exited, nil
	case "dead":
		return common.Dead, nil
	default:
		return common.Unknown, fmt.Errorf("unknown container status: %s", ctInfo.State.Status)
	}
}

// PS lists all running containers along with their details
//
// PS returns a slice of ContainerInfo structs, each representing one
// running container. ContainerInfo struct contains the ID, Name, Image, Command, Created,
// Ports, and Status of the container. If an error occurs while communicating with the Docker API,
// function returns nil and the error.
func (d *DockerManager) PS() ([]ContainerInfo, error) {
	log.Debugf("Listing containers")
	containerList, err := d.dockerClient.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		return nil, err
	}
	ctInfo := make([]ContainerInfo, len(containerList))
	for i, ct := range containerList {
		ctInfo[i] = ContainerInfo{
			ID:      ct.ID,
			Names:   ct.Names,
			Image:   ct.Image,
			Command: ct.Command,
			Created: ct.Created,
			Ports:   convertPorts(ct.Ports),
			Status:  ct.Status,
		}
	}
	return ctInfo, nil
}

// ContainerIP returns the IP address of the specified container
//
// The function takes a container ID or name as input and returns its IP address
// as a string. If an error occurs while inspecting the container or if the container
// does not have any networks configured, function returns an empty string and an error.
// Possible errors that may occur are returned as error types defined in this package.
func (d *DockerManager) ContainerIP(container string) (string, error) {
	log.Debugf("Getting container's IP: %s", container)
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), container)
	if err != nil {
		return "", err
	}
	networks := ctInfo.NetworkSettings.Networks
	if len(networks) == 0 {
		return "", fmt.Errorf("%w: in %s", ErrNetworksNotFound, container)
	}
	var ipAddress string
	for _, network := range networks {
		ipAddress = network.IPAddress
		break // Use the first IP address found.
	}
	return ipAddress, nil
}
