package docker

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/docker/mocks"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

// Image tests

func TestImageFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	expectedImage := "expected-image"

	dockerClient.EXPECT().
		ContainerInspect(context.Background(), "eigen").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				Image: expectedImage,
			},
		}, nil)

	dockerManager := NewDockerManager(dockerClient)
	image, err := dockerManager.Image("eigen")
	assert.Nil(t, err)
	assert.Equal(t, expectedImage, image)
}

func TestImageNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	expectedError := errdefs.NotFound(fmt.Errorf("error"))

	dockerClient.EXPECT().
		ContainerInspect(context.Background(), "eigen").
		Return(types.ContainerJSON{}, expectedError)

	dockerManager := NewDockerManager(dockerClient)
	image, err := dockerManager.Image("eigen")
	assert.ErrorIs(t, err, expectedError)
	assert.True(t, errdefs.IsNotFound(err))
	assert.Equal(t, "", image)
}

func ExampleDockerManager_Image() {
	// Create a new Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer dockerClient.Close()

	// Create a new dockerManager
	dm := NewDockerManager(dockerClient)

	// Get the image name of a running container
	_, err = dm.Image("myContainer")
	if err != nil {
		log.Error(err)
	}
}

// Start tests

func TestStartError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	dockerClient.EXPECT().
		ContainerStart(context.Background(), "eigen", gomock.Any()).
		Return(errors.New("error"))

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Start("eigen")
	assert.ErrorIs(t, err, ErrStartingContainer)
}

func TestStartWithoutError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	dockerClient.EXPECT().
		ContainerStart(context.Background(), "eigen", gomock.Any()).
		Return(nil)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Start("eigen")
	assert.Nil(t, err)
}

// Stop tests

func TestStopContainerNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	expectedError := errdefs.NotFound(errors.New("error"))

	dockerClient.EXPECT().
		ContainerInspect(context.Background(), "eigen").
		Return(types.ContainerJSON{}, expectedError)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Stop("eigen")
	assert.Nil(t, err)
}

func TestStopError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	expectedError := errors.New("error")

	dockerClient.EXPECT().
		ContainerInspect(context.Background(), "eigen").
		Return(types.ContainerJSON{}, expectedError)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Stop("eigen")
	assert.ErrorIs(t, err, expectedError)
}

func TestStopContainerAlreadyStopped(t *testing.T) {
	// Silence logger
	log.SetOutput(io.Discard)

	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	expectedError := errors.New("error")
	eigenCtId := "eigenctid"

	dockerClient.EXPECT().
		ContainerInspect(context.Background(), "eigen").
		Return(types.ContainerJSON{
			ContainerJSONBase: &types.ContainerJSONBase{
				ID: eigenCtId,
				State: &types.ContainerState{
					Running: true,
				},
			},
		}, nil)
	dockerClient.EXPECT().
		ContainerStop(context.Background(), eigenCtId, gomock.Any()).
		Return(expectedError)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Stop("eigen")
	assert.ErrorIs(t, err, ErrStoppingContainer)
}

func TestStopContainer(t *testing.T) {
	// Silence logger
	log.SetOutput(io.Discard)

	tests := []struct {
		name           string
		containerState *types.ContainerState
	}{
		{
			name: "Running status success",
			containerState: &types.ContainerState{
				Running: true,
			},
		},
		{
			name: "Restarting status success",
			containerState: &types.ContainerState{
				Restarting: true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()

			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			defer ctrl.Finish()

			eigenCtId := "eigenctid"
			dockerClient.EXPECT().
				ContainerInspect(context.Background(), "eigen").
				Return(types.ContainerJSON{
					ContainerJSONBase: &types.ContainerJSONBase{
						ID:    eigenCtId,
						State: tt.containerState,
					},
				}, nil)
			dockerClient.EXPECT().
				ContainerStop(context.Background(), eigenCtId, gomock.Any()).
				Return(nil)

			dockerManager := NewDockerManager(dockerClient)
			err := dockerManager.Stop("eigen")
			assert.Nil(t, err)
		})
	}
}

// ContainerID tests

func TestContainerId(t *testing.T) {
	ctName := "container-name"
	tests := []struct {
		name       string
		containers []types.Container
		wantId     string
		err        error
	}{
		{
			name: "container found",
			containers: []types.Container{
				{
					ID:    "other-id",
					Names: []string{"other-name"},
				},
				{
					ID:    "container-id",
					Names: []string{"/" + ctName},
				},
			},
			wantId: "container-id",
			err:    nil,
		},
		{
			name:       "container not found, no containers",
			containers: []types.Container{},
			wantId:     "",
			err:        ErrContainerNotFound,
		},
		{
			name: "container found, no exact match",
			containers: []types.Container{
				{
					ID:    "other-id",
					Names: []string{"other-name"},
				},
				{
					ID:    "container-id",
					Names: []string{ctName + "-2", ctName + "-3"},
				},
			},
			wantId: "",
			err:    ErrContainerNotFound,
		},
	}
	for _, tt := range tests {
		ctrl := gomock.NewController(t)
		dockerClient := mocks.NewMockAPIClient(ctrl)
		defer ctrl.Finish()

		dockerClient.EXPECT().
			ContainerList(gomock.Any(), types.ContainerListOptions{
				All:     true,
				Filters: filters.NewArgs(filters.Arg("name", ctName)),
			}).
			Return(tt.containers, nil)
		dockerManager := NewDockerManager(dockerClient)
		id, err := dockerManager.ContainerID(ctName)
		assert.ErrorIs(t, err, tt.err)
		assert.Equal(t, tt.wantId, id)

	}
}

func TestContainerIdError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	wantErr := errors.New("error")
	containerName := "container-name"

	dockerClient.EXPECT().
		ContainerList(gomock.Any(), types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", containerName)),
		}).
		Return(nil, wantErr)

	dockerManager := NewDockerManager(dockerClient)
	id, err := dockerManager.ContainerID(containerName)
	assert.ErrorIs(t, err, wantErr)
	assert.Equal(t, "", id)
}

func TestContainerIdNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	containerName := "container-name"

	dockerClient.EXPECT().
		ContainerList(gomock.Any(), types.ContainerListOptions{
			All:     true,
			Filters: filters.NewArgs(filters.Arg("name", containerName)),
		}).
		Return(make([]types.Container, 0), nil)

	dockerManager := NewDockerManager(dockerClient)
	id, err := dockerManager.ContainerID(containerName)
	assert.ErrorIs(t, err, ErrContainerNotFound)
	assert.Equal(t, "", id)
}

// Pull tests

func TestPullError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	wantErr := errors.New("error")
	imageName := "eigen:latest"

	dockerClient.EXPECT().
		ImagePull(gomock.Any(), imageName, gomock.Any()).
		Return(nil, wantErr)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Pull(imageName)
	assert.ErrorIs(t, err, wantErr)
}

func TestPull(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	imageName := "eigen:latest"

	dockerClient.EXPECT().
		ImagePull(gomock.Any(), imageName, gomock.Any()).
		Return(nil, nil)

	dockerManager := NewDockerManager(dockerClient)
	err := dockerManager.Pull(imageName)
	assert.Nil(t, err)
}

// ContainerLogs tests

func TestContainerLogsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	wantErr := errors.New("error")
	containerName := "eigen"

	dockerClient.EXPECT().
		ContainerLogs(gomock.Any(), containerName, gomock.Any()).
		Return(nil, wantErr)

	dockerManager := NewDockerManager(dockerClient)
	logs, err := dockerManager.ContainerLogs(containerName)
	assert.ErrorIs(t, err, wantErr)
	assert.Empty(t, logs)
}

type mockReadCloser struct {
	io.Reader
}

func (mockReadCloser) Close() error { return nil }

// Ensure mockReadCloser implements io.ReadCloser
var _ io.ReadCloser = (*mockReadCloser)(nil)

func TestContainerLogs(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	containerName := "eigen"
	wantLogs := "logs"

	logReader := mockReadCloser{Reader: strings.NewReader(wantLogs)}

	dockerClient.EXPECT().
		ContainerLogs(gomock.Any(), containerName, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     false,
		}).
		Return(logReader, nil)

	dockerManager := NewDockerManager(dockerClient)
	logs, err := dockerManager.ContainerLogs(containerName)
	assert.Nil(t, err)
	assert.Equal(t, wantLogs, logs)
}

// Wait tests
func TestWaitErrCh(t *testing.T) {
	ctrl := gomock.NewController(t)
	dockerClient := mocks.NewMockAPIClient(ctrl)
	defer ctrl.Finish()

	waitCh := time.After(3 * time.Second)
	wantErr := errors.New("error")
	wantErrCh := make(chan error, 1)
	wantErrCh <- wantErr

	dockerClient.EXPECT().
		ContainerWait(context.Background(), "eigen", gomock.Any()).
		Return(make(chan container.WaitResponse), wantErrCh)

	dockerManager := NewDockerManager(dockerClient)
	exitCh, errCh := dockerManager.Wait("eigen", WaitConditionNextExit)
	select {
	case <-waitCh:
		t.Fatal("err channel timeout")
	case <-exitCh:
		t.Fatal("unexpected value from exit channel")
	case err := <-errCh:
		assert.ErrorIs(t, err, wantErr)
	}
}

func TestWaitExitCh(t *testing.T) {
	tests := []struct {
		name     string
		response container.WaitResponse
	}{
		{
			name: "response without error",
			response: container.WaitResponse{
				StatusCode: 0,
			},
		},
		{
			name: "response with error",
			response: container.WaitResponse{
				StatusCode: 0,
				Error: &container.WaitExitError{
					Message: "error",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			defer ctrl.Finish()

			waitCh := time.After(3 * time.Second)
			wantWaitCh := make(chan container.WaitResponse, 1)
			wantWaitCh <- tt.response
			close(wantWaitCh)

			dockerClient.EXPECT().
				ContainerWait(context.Background(), "eigen", gomock.Any()).
				Return(wantWaitCh, make(chan error))

			dockerManager := NewDockerManager(dockerClient)
			exitCh, errCh := dockerManager.Wait("eigen", WaitConditionNextExit)
			select {
			case <-waitCh:
				t.Fatal("exit channel timeout")
			case exit := <-exitCh:
				assert.Equal(t, tt.response.StatusCode, exit.StatusCode)
				if tt.response.Error == nil {
					assert.Nil(t, exit.Error)
				} else {
					assert.Equal(t, tt.response.Error.Message, exit.Error.Message)
				}
			case <-errCh:
				t.Fatal("unexpected value from error channel")
			}
		})
	}
}

func TestContainerStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		want     common.Status
		errorMsg string
		wantErr  bool
	}{
		{
			name:   "container created",
			status: "created",
			want:   common.Created,
		},
		{
			name:   "container running",
			status: "running",
			want:   common.Running,
		},
		{
			name:   "container paused",
			status: "paused",
			want:   common.Paused,
		},
		{
			name:   "container restarting",
			status: "restarting",
			want:   common.Restarting,
		},
		{
			name:   "container removing",
			status: "removing",
			want:   common.Removing,
		},
		{
			name:   "container exited",
			status: "exited",
			want:   common.Exited,
		},
		{
			name:   "container dead",
			status: "dead",
			want:   common.Dead,
		},
		{
			name:     "container unknown",
			status:   "unknown",
			want:     common.Unknown,
			errorMsg: "unknown container status: unknown",
		},
		{
			name:     "bad container",
			want:     common.Unknown,
			wantErr:  true,
			errorMsg: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			defer ctrl.Finish()

			container := types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					State: &types.ContainerState{
						Status: tt.status,
					},
				},
			}

			if tt.wantErr {
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container.ID).
					Return(container, errors.New("error"))
			} else {
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container.ID).
					Return(container, nil)
			}

			dockerManager := NewDockerManager(dockerClient)
			status, err := dockerManager.ContainerStatus(container.ID)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.Equal(t, tt.want, status)
			}
		})
	}
}
