package docker

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/NethermindEth/eigen-wiz/internal/utils"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func NewDockerManager(dockerClient client.APIClient) dockerManager {
	return dockerManager{dockerClient}
}

// DockerManager is an interface to the Docker Daemon for managing docker containers
type dockerManager struct {
	dockerClient client.APIClient
}

// Returns the image name of a container
func (d *dockerManager) Image(containerName string) (img string, err error) {
	var ctInfo types.ContainerJSON
	if ctInfo, err = d.dockerClient.ContainerInspect(context.Background(), containerName); err != nil {
		return "", err
	}
	return ctInfo.Image, nil
}

// Starts a container
func (d *dockerManager) Start(containerName string) error {
	if err := d.dockerClient.ContainerStart(context.Background(), containerName, types.ContainerStartOptions{}); err != nil {
		return fmt.Errorf("%w %s: %s", ErrStartingContainer, containerName, err)
	}
	return nil
}

// Stops a container
func (d *dockerManager) Stop(containerName string) error {
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), containerName)
	if err != nil {
		if client.IsErrNotFound(err) {
			return nil
		}
		return err
	}
	if ctInfo.State.Running {
		log.Infof("Stopping service: %s, currently on %s status", containerName, ctInfo.State.Status)
		timeout := 5 * int(time.Minute)
		if err := d.dockerClient.ContainerStop(context.Background(), ctInfo.ID, container.StopOptions{
			Timeout: &timeout,
		}); err != nil {
			return fmt.Errorf("%w %s: %s", ErrStoppingContainer, containerName, err)
		}
	}
	return nil
}

// Returns the container id of a container
func (d *dockerManager) ContainerID(containerName string) (string, error) {
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

// Pulls an image
func (d *dockerManager) Pull(image string) error {
	log.Debugf("Pulling image: %s", image)
	_, err := d.dockerClient.ImagePull(context.Background(), image, types.ImagePullOptions{})
	return err
}

// Returns the logs of a container
func (d *dockerManager) ContainerLogs(containerID string) (string, error) {
	logReader, err := d.dockerClient.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return "", err
	}
	defer logReader.Close()

	logs, err := io.ReadAll(logReader)
	log.Errorf("%p", &logReader)
	if err == nil {
		log.Debugf("Container logs: %s", string(logs))
	}
	return string(logs), err
}

// Waits for a container to reach a certain condition
func (d *dockerManager) Wait(service string, condition container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	return d.dockerClient.ContainerWait(context.Background(), service, condition)
}
