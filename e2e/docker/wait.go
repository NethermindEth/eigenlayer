package docker

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/docker/docker/client"
)

func WaitUntilRunning(containerID string, timeout time.Duration) error {
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}
	defer dockerClient.Close()

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	b := backoff.WithContext(backoff.NewExponentialBackOff(), ctx)

	notRunningErr := errors.New("container is not running")
	return backoff.Retry(func() error {
		response, err := dockerClient.ContainerInspect(ctx, containerID)
		if err != nil {
			return backoff.Permanent(err)
		}
		if response.State.Running {
			return nil
		} else {
			return notRunningErr
		}
	}, b)
}
