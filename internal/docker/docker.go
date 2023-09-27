package docker

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"path/filepath"
	"sync"
	"time"

	"github.com/docker/docker/api/types"
	dockerCt "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	log "github.com/sirupsen/logrus"

	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/internal/utils"
)

const NetworkHost = "host"

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

type ContainerLogsMergedOptions struct {
	Follow     bool
	Since      string
	Until      string
	Timestamps bool
	Tail       string
}

// ContainerLogsMerged retrieves the merge of all logs of specified Docker containers.
// Spin up a goroutine for each container and write the logs to the provided io.Writer,
// and wait for all goroutines to finish. To cancel the process, use the provided context.
func (d *DockerManager) ContainerLogsMerged(ctx context.Context, w io.Writer, services map[string]string, opts ContainerLogsMergedOptions) error {
	wLock := new(sync.Mutex)
	waitGroup := new(sync.WaitGroup)
	errCh := make(chan error, len(services))

	for serviceName, id := range services {
		waitGroup.Add(1)
		go func(serviceName, id string, out chan error) {
			defer waitGroup.Done()

			logReader, err := d.dockerClient.ContainerLogs(ctx, id, types.ContainerLogsOptions{
				ShowStdout: true,
				ShowStderr: true,
				Follow:     opts.Follow,
				Since:      opts.Since,
				Until:      opts.Until,
				Timestamps: opts.Timestamps,
				Tail:       opts.Tail,
			})
			if err != nil {
				out <- fmt.Errorf("error getting logs for %s: %w", serviceName, err)
				return
			}
			defer logReader.Close()

			scanner := bufio.NewScanner(logReader)
			for scanner.Scan() {
				wLock.Lock()
				w.Write([]byte(serviceName + ": "))
				scannerBytes := scanner.Bytes()

				/* Remove the first 8 bytes, which are the logs HEADER from Docker.
				Each log line has the format <HEADER><PAYLOAD>. The HEADER is 8 bytes long:
				- First byte indicates the stream type (stdout or stderr)
				- Next 3 bytes unused, padding
				- Next 4 bytes indicate the length of the log payload
				The PAYLOAD is the actual log line.*/
				logPayload := scannerBytes[8:]

				w.Write(logPayload)
				w.Write([]byte("\n"))
				wLock.Unlock()
			}
			out <- nil
		}(serviceName, id, errCh)
	}
	waitGroup.Wait()
	// Build error
	var err error
	for len(errCh) > 0 {
		if e := <-errCh; e != nil {
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("%w. %w", err, e)
			}
		}
	}
	close(errCh)
	return err
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

// ContainerNetworks returns the networks of the specified container
func (d *DockerManager) ContainerNetworks(container string) ([]string, error) {
	log.Debugf("Getting container's networks: %s", container)
	ctInfo, err := d.dockerClient.ContainerInspect(context.Background(), container)
	if err != nil {
		return nil, err
	}
	networks := ctInfo.NetworkSettings.Networks
	// TODO: Consider whether returning an error if the container has no networks is a good idea.
	if len(networks) == 0 {
		return nil, fmt.Errorf("%w: in %s", ErrNetworksNotFound, container)
	}
	var networkNames []string
	for network := range networks {
		networkNames = append(networkNames, network)
	}
	return networkNames, nil
}

// NetworkConnect connects a container to a network
func (d *DockerManager) NetworkConnect(container, network string) error {
	log.Debugf("Connecting container %s to network %s", container, network)
	return d.dockerClient.NetworkConnect(context.Background(), network, container, nil)
}

// NetworkDisconnect disconnects a container from a network
func (d *DockerManager) NetworkDisconnect(container, network string) error {
	log.Debugf("Disconnecting container %s from network %s", container, network)
	return d.dockerClient.NetworkDisconnect(context.Background(), network, container, false)
}

// BuildFromURL build an image from a Git repository URI or HTTP/HTTPS context URI.
func (d *DockerManager) BuildImageFromURI(remote string, tag string, buildArgs map[string]*string) (err error) {
	log.Debugf("Building image from %s", remote)
	buildResult, err := d.dockerClient.ImageBuild(context.Background(), nil, types.ImageBuildOptions{
		RemoteContext: remote,
		BuildArgs:     buildArgs,
		Tags:          []string{tag},
		Remove:        true,
		ForceRemove:   true,
	})
	if err != nil {
		return err
	}
	defer buildResult.Body.Close()

	loadResult, err := d.dockerClient.ImageLoad(context.Background(), buildResult.Body, true)
	if err != nil {
		return err
	}
	defer loadResult.Body.Close()
	return nil
}

// LoadImageContext loads an image context from a local path.
func (d *DockerManager) LoadImageContext(path string) (io.ReadCloser, error) {
	log.Debug("Loading context from ", path)
	if !filepath.IsAbs(path) {
		return nil, fmt.Errorf("path must be absolute")
	}
	return archive.TarWithOptions(path, &archive.TarOptions{})
}

// TODO: [REFACTOR] Remove this function if it's not used.
// BuildFromLocalPath build a docker image from a local path.
// func (d *DockerManager) BuildImageFromContext(ctx io.ReadCloser, tag string, buildArgs map[string]*string) (err error) {
// 	log.Debugf("Building image from context")
// 	// Build plugin image
// 	buildResult, err := d.dockerClient.ImageBuild(context.Background(), ctx, types.ImageBuildOptions{
// 		BuildArgs:   buildArgs,
// 		Tags:        []string{tag},
// 		Remove:      true,
// 		ForceRemove: true,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	defer buildResult.Body.Close()
// 	// Save image
// 	loadResult, err := d.dockerClient.ImageLoad(context.Background(), buildResult.Body, true)
// 	if err != nil {
// 		return err
// 	}
// 	defer loadResult.Body.Close()
// 	return nil
// }

type VolumeType string

const (
	VolumeTypeBind   VolumeType = "bind"
	VolumeTypeVolume VolumeType = "volume"
)

type Mount struct {
	Type   VolumeType
	Source string
	Target string
}

func (m Mount) mount() mount.Mount {
	return mount.Mount{
		Type:   mount.Type(m.Type),
		Source: m.Source,
		Target: m.Target,
	}
}

// ImageRemove removes a specified Docker image.
func (d *DockerManager) ImageRemove(image string) error {
	log.Debugf("Removing image %s", image)
	_, err := d.dockerClient.ImageRemove(context.Background(), image, types.ImageRemoveOptions{
		PruneChildren: true,
	})
	return err
}

type RunOptions struct {
	Network     string
	Args        []string
	Mounts      []Mount
	VolumesFrom []string
	AutoRemove  bool
}

// Run is a method of DockerManager that handles running a Docker container from an image.
// It creates the container from the specified image with the provided command arguments,
// connects the created container to the specified network, then starts the container.
//
// After the container starts, the function waits for the container to exit.
// During the waiting process, it also listens for errors from the container.
// If an error is received, it prints the container logs and returns the error.
func (d *DockerManager) Run(image string, options RunOptions) (err error) {
	log.Debugf("Creating container from image %s", image)
	// Build mounts
	hostConfig := &dockerCt.HostConfig{
		VolumesFrom: options.VolumesFrom,
		AutoRemove:  options.AutoRemove,
	}
	for _, mount := range options.Mounts {
		switch mount.Type {
		case VolumeTypeBind, VolumeTypeVolume:
			hostConfig.Mounts = append(hostConfig.Mounts, mount.mount())
		default:
			return fmt.Errorf("unknown mount type: %s", mount.Type)
		}
	}
	// Create container
	createResponse, err := d.dockerClient.ContainerCreate(context.Background(),
		&dockerCt.Config{Image: image, Cmd: options.Args},
		hostConfig, nil, nil, "")
	if err != nil {
		return err
	}

	// Ensure the container is removed after use
	defer func() {
		log.Debugf("Removing container %s", createResponse.ID)
		removeErr := d.dockerClient.ContainerRemove(context.Background(), createResponse.ID, types.ContainerRemoveOptions{})
		if removeErr != nil {
			// If the main function did not return an error, but the deferred function did,
			// the deferred function's error is returned.
			if err == nil {
				err = removeErr
			} else {
				log.Errorf("Error removing container %s: %v", createResponse.ID, removeErr)
			}
		}
	}()

	if options.Network != "" && options.Network != NetworkHost {
		log.Debugf("Connecting container %s to network %s", createResponse.ID, options.Network)
		err = d.NetworkConnect(createResponse.ID, options.Network)
		if err != nil {
			return err
		}
	}
	waitChn, errChn := d.dockerClient.ContainerWait(context.Background(), createResponse.ID, dockerCt.WaitConditionNextExit)
	log.Debugf("Starting container %s", createResponse.ID)
	err = d.dockerClient.ContainerStart(context.Background(), createResponse.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	log.Debugf("Waiting for container to exit")
	for {
		select {
		case err := <-errChn:
			if err != nil {
				log.Debugf("Error while waiting for container to exit")
				return fmt.Errorf("error waiting for container %s: %v. container logs: %s", createResponse.ID, err, containerLogs(d.dockerClient, createResponse.ID))
			}
		case wait := <-waitChn:
			log.Debugf("Container exited with status %d", wait.StatusCode)
			if wait.StatusCode != 0 {
				return fmt.Errorf("unexpected exit code %d for container %s. container logs: %s", wait.StatusCode, createResponse.ID, containerLogs(d.dockerClient, createResponse.ID))
			} else {
				log.Info(containerLogs(d.dockerClient, createResponse.ID))
			}
			return nil
		}
	}
}

func containerLogs(dockerClient client.APIClient, containerID string) string {
	logsReader, err := dockerClient.ContainerLogs(context.Background(), containerID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		log.Errorf("Error getting container logs: %v", err)
		return ""
	}
	defer logsReader.Close()
	logs, err := io.ReadAll(logsReader)
	if err != nil {
		log.Errorf("Error reading container logs: %v", err)
		return ""
	}
	return string(logs)
}
