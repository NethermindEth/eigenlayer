package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/errdefs"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/docker/mocks"
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

func TestPS(t *testing.T) {
	tests := []struct {
		name       string
		containers []types.Container
		want       []ContainerInfo
		wantErr    error
	}{
		{
			name: "one container found",
			containers: []types.Container{
				{
					ID:      "container-id1",
					Names:   []string{"/name1"},
					Image:   "image1",
					ImageID: "image-id1",
					Command: "command1",
					Created: 1234,
					Ports: []types.Port{
						{
							IP:          "127.0.0.1",
							PrivatePort: 3000,
							PublicPort:  3080,
						},
					},
					Status: "running",
					State:  "state1",
				},
			},
			want: []ContainerInfo{
				{
					ID:      "container-id1",
					Names:   []string{"/name1"},
					Image:   "image1",
					Command: "command1",
					Created: 1234,
					Ports: []Port{
						{
							IP:          "127.0.0.1",
							PrivatePort: 3000,
							PublicPort:  3080,
						},
					},
					Status: "running",
				},
			},
			wantErr: nil,
		},
		{
			name: "many containers found",
			containers: []types.Container{
				{
					ID:      "container-id1",
					Names:   []string{"/name1"},
					Image:   "image1",
					ImageID: "image-id1",
					Command: "command1",
					Created: 1234,
					Ports: []types.Port{
						{
							IP:          "127.0.0.1",
							PrivatePort: 3000,
							PublicPort:  3080,
						},
					},
					Status: "running",
					State:  "state1",
				},
				{
					ID:      "container-id2",
					Names:   []string{"/name2"},
					Image:   "image2",
					ImageID: "image-id2",
					Command: "command2",
					Created: 5678,
					Ports: []types.Port{
						{
							IP:          "127.0.0.10",
							PrivatePort: 4000,
							PublicPort:  4080,
						},
					},
					Status: "running",
					State:  "state2",
				},
			},
			want: []ContainerInfo{
				{
					ID:      "container-id1",
					Names:   []string{"/name1"},
					Image:   "image1",
					Command: "command1",
					Created: 1234,
					Ports: []Port{
						{
							IP:          "127.0.0.1",
							PrivatePort: 3000,
							PublicPort:  3080,
						},
					},
					Status: "running",
				},
				{
					ID:      "container-id2",
					Names:   []string{"/name2"},
					Image:   "image2",
					Command: "command2",
					Created: 5678,
					Ports: []Port{
						{
							IP:          "127.0.0.10",
							PrivatePort: 4000,
							PublicPort:  4080,
						},
					},
					Status: "running",
				},
			},
		},
		{
			name:       "none containers",
			containers: []types.Container{},
			want:       []ContainerInfo{},
		},
		{
			name:       "returning error",
			containers: []types.Container{},
			want:       nil,
			wantErr:    errors.New("error listing containers"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			defer ctrl.Finish()

			dockerClient.EXPECT().ContainerList(gomock.Any(), types.ContainerListOptions{}).Return(tt.containers, tt.wantErr)

			dockerManager := NewDockerManager(dockerClient)
			got, err := dockerManager.PS()
			assert.ErrorIs(t, err, tt.wantErr, "Unexpected error returned")
			assert.Equal(t, tt.want, got, "Expected containers does not match with containers obtained.")
		})
	}
}

func TestContainerIP(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		response    types.ContainerJSON
		want        string
		wantErr     error
		expectedErr error
	}{
		{
			name: "returning error",
			arg:  "eigen",
			response: types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					ID: "container-Id",
					State: &types.ContainerState{
						Running: true,
					},
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"eigen": {
							IPAddress: "127.0.0.1",
						},
					},
				},
			},
			want:    "",
			wantErr: errors.New("error inspecting container"),
		},
		{
			name: "checking empty networks",
			arg:  "sedge-network",
			response: types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					ID: "container-Id",
					State: &types.ContainerState{
						Running: true,
					},
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{},
				},
			},
			want:        "",
			expectedErr: ErrNetworksNotFound,
		},
		{
			name: "returning correct IP",
			arg:  "sedge-network",
			response: types.ContainerJSON{
				ContainerJSONBase: &types.ContainerJSONBase{
					ID: "container-Id",
					State: &types.ContainerState{
						Running: true,
					},
				},
				NetworkSettings: &types.NetworkSettings{
					Networks: map[string]*network.EndpointSettings{
						"eigen": {
							IPAddress: "127.0.0.1",
						},
					},
				},
			},
			want: "127.0.0.1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			defer ctrl.Finish()
			dockerClient.EXPECT().
				ContainerInspect(context.Background(), tt.arg).
				Return(tt.response, tt.wantErr)

			dockerManager := NewDockerManager(dockerClient)
			got, err := dockerManager.ContainerIP(tt.arg)
			if tt.wantErr == nil {
				assert.ErrorIs(t, err, tt.expectedErr, "Unexpected error returned")
			} else {
				assert.ErrorIs(t, err, tt.wantErr, "Unexpected error returned")
			}
			assert.Equal(t, tt.want, got, "Expected container IP does not match with IP obtained.")
		})
	}
}

func TestContainerNetworks(t *testing.T) {
	tests := []struct {
		name      string
		mocker    func(t *testing.T, container string) *mocks.MockAPIClient
		container string
		want      []string
		wantErr   bool
	}{
		{
			name: "ok, 1 network",
			mocker: func(t *testing.T, container string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container).
					Return(types.ContainerJSON{
						NetworkSettings: &types.NetworkSettings{
							Networks: map[string]*network.EndpointSettings{
								"eigen": {
									IPAddress: "127.0.0.1",
								},
							},
						},
					}, nil)
				return dockerClient
			},
			container: "container-Id1",
			want:      []string{"eigen"},
			wantErr:   false,
		},
		{
			name: "ok, 2 networks",
			mocker: func(t *testing.T, container string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container).
					Return(types.ContainerJSON{
						NetworkSettings: &types.NetworkSettings{
							Networks: map[string]*network.EndpointSettings{
								"eigen": {
									IPAddress: "168.0.0.1",
								},
								"eigen2": {
									IPAddress: "168.0.0.2",
								},
							},
						},
					}, nil)
				return dockerClient
			},
			container: "container-Id2",
			want:      []string{"eigen", "eigen2"},
			wantErr:   false,
		},
		{
			name: "empty networks",
			mocker: func(t *testing.T, container string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container).
					Return(types.ContainerJSON{
						NetworkSettings: &types.NetworkSettings{
							Networks: map[string]*network.EndpointSettings{},
						},
					}, nil)
				return dockerClient
			},
			container: "container-Id3",
			wantErr:   true,
		},
		{
			name: "error inspecting container",
			mocker: func(t *testing.T, container string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					ContainerInspect(context.Background(), container).
					Return(types.ContainerJSON{}, errors.New("error inspecting container"))
				return dockerClient
			},
			container: "container-Id4",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerClient := tt.mocker(t, tt.container)
			dockerManager := NewDockerManager(dockerClient)
			got, err := dockerManager.ContainerNetworks(tt.container)
			if tt.wantErr {
				assert.Error(t, err, "Expected error not returned")
			} else {
				assert.NoError(t, err, "Unexpected error returned")
			}
			assert.Len(t, got, len(tt.want), "Expected container networks does not match with networks obtained.")
			for _, network := range tt.want {
				assert.Contains(t, got, network, "Expected container networks does not match with networks obtained.")
			}
		})
	}
}

func TestNetworkConnect(t *testing.T) {
	tests := []struct {
		name      string
		mocker    func(t *testing.T, container, network string) *mocks.MockAPIClient
		container string
		network   string
		wantErr   bool
	}{
		{
			name: "ok",
			mocker: func(t *testing.T, container, network string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					NetworkConnect(context.Background(), network, container, nil).
					Return(nil)
				return dockerClient
			},
			container: "container-Id1",
			network:   "eigen",
		},
		{
			name: "error connecting container to network",
			mocker: func(t *testing.T, container, network string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					NetworkConnect(context.Background(), network, container, nil).
					Return(errors.New("error connecting container to network"))
				return dockerClient
			},
			container: "container-Id2",
			network:   "eigen",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerClient := tt.mocker(t, tt.container, tt.network)
			dockerManager := NewDockerManager(dockerClient)
			err := dockerManager.NetworkConnect(tt.container, tt.network)
			if tt.wantErr {
				assert.Error(t, err, "Expected error not returned")
			} else {
				assert.NoError(t, err, "Unexpected error returned")
			}
		})
	}
}

func TestNetworkDisconnect(t *testing.T) {
	tests := []struct {
		name      string
		mocker    func(t *testing.T, container, network string) *mocks.MockAPIClient
		container string
		network   string
		wantErr   bool
	}{
		{
			name: "ok",
			mocker: func(t *testing.T, container, network string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					NetworkDisconnect(context.Background(), network, container, false).
					Return(nil)
				return dockerClient
			},
			container: "container-Id1",
			network:   "eigen",
		},
		{
			name: "error disconnecting container from network",
			mocker: func(t *testing.T, container, network string) *mocks.MockAPIClient {
				ctrl := gomock.NewController(t)
				dockerClient := mocks.NewMockAPIClient(ctrl)
				dockerClient.EXPECT().
					NetworkDisconnect(context.Background(), network, container, false).
					Return(errors.New("error disconnecting container from network"))
				return dockerClient
			},
			container: "container-Id2",
			network:   "eigen",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dockerClient := tt.mocker(t, tt.container, tt.network)
			dockerManager := NewDockerManager(dockerClient)
			err := dockerManager.NetworkDisconnect(tt.container, tt.network)
			if tt.wantErr {
				assert.Error(t, err, "Expected error not returned")
			} else {
				assert.NoError(t, err, "Unexpected error returned")
			}
		})
	}
}

func TestBuildFromURI(t *testing.T) {
	tests := []struct {
		name          string
		remote        string
		tag           string
		setup         func(*mocks.MockAPIClient)
		expectedError error
	}{
		{
			name:   "success",
			remote: "https://github.com/NethermindEth/mock-avs#main:plugin",
			tag:    "mock-avs-plugin",
			setup: func(dockerClient *mocks.MockAPIClient) {
				buildBody := io.NopCloser(bytes.NewReader([]byte{}))
				buildResponse := types.ImageBuildResponse{
					Body: buildBody,
				}
				loadBody := io.NopCloser(bytes.NewReader([]byte{}))
				loadResponse := types.ImageLoadResponse{
					Body: loadBody,
				}
				dockerClient.EXPECT().ImageBuild(context.Background(), nil, types.ImageBuildOptions{
					RemoteContext: "https://github.com/NethermindEth/mock-avs#main:plugin",
					Tags:          []string{"mock-avs-plugin"},
					Remove:        true,
					ForceRemove:   true,
				}).Return(buildResponse, nil)
				dockerClient.EXPECT().ImageLoad(context.Background(), buildResponse.Body, true).Return(loadResponse, nil)
			},
			expectedError: nil,
		},
		{
			name:   "build error",
			remote: "https://github.com/NethermindEth/mock-avs#main:plugin",
			tag:    "mock-avs-plugin",
			setup: func(dockerClient *mocks.MockAPIClient) {
				dockerClient.EXPECT().ImageBuild(context.Background(), nil, types.ImageBuildOptions{
					RemoteContext: "https://github.com/NethermindEth/mock-avs#main:plugin",
					Tags:          []string{"mock-avs-plugin"},
					Remove:        true,
					ForceRemove:   true,
				}).Return(types.ImageBuildResponse{}, errors.New("build error"))
			},
			expectedError: errors.New("build error"),
		},
		{
			name:   "load error",
			remote: "https://github.com/orgname/avs#main:plugin",
			tag:    "orgname-avs-plugin",
			setup: func(dockerClient *mocks.MockAPIClient) {
				buildBody := io.NopCloser(bytes.NewReader([]byte{}))
				buildResponse := types.ImageBuildResponse{
					Body: buildBody,
				}
				dockerClient.EXPECT().ImageBuild(context.Background(), nil, types.ImageBuildOptions{
					RemoteContext: "https://github.com/orgname/avs#main:plugin",
					Tags:          []string{"orgname-avs-plugin"},
					Remove:        true,
					ForceRemove:   true,
				}).Return(buildResponse, nil)
				dockerClient.EXPECT().ImageLoad(context.Background(), buildResponse.Body, true).Return(types.ImageLoadResponse{}, errors.New("load error"))
			},
			expectedError: errors.New("load error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			dockerClient := mocks.NewMockAPIClient(ctrl)
			tt.setup(dockerClient)
			defer ctrl.Finish()

			dockerManager := NewDockerManager(dockerClient)
			err := dockerManager.BuildFromURI(tt.remote, tt.tag)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDockerManager_Run(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*gomock.Controller) *mocks.MockAPIClient
		image         string
		network       string
		args          []string
		expectedError error
	}{
		{
			name: "Run successful",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)

				// Create channels
				waitCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				// Write to one of the channels
				waitCh <- container.WaitResponse{StatusCode: 0}

				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerWait(gomock.Any(), "containerID", gomock.Any()).Return(waitCh, errCh),
					dockerClient.EXPECT().ContainerStart(gomock.Any(), "containerID", gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerLogs(gomock.Any(), "containerID", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}).Return(io.NopCloser(bytes.NewBuffer([]byte{})), nil),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(nil),
				)
				return dockerClient
			},
			image:   "my-image",
			network: "my-network",
			args:    []string{"arg1", "arg2"},
		},
		{
			name: "Container create error",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)
				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{}, errors.New("creation error")),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: errors.New("creation error"),
		},
		{
			name: "Container remove error",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)

				// Create channels
				waitCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				// Write to one of the channels
				waitCh <- container.WaitResponse{StatusCode: 0}

				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerWait(gomock.Any(), "containerID", gomock.Any()).Return(waitCh, errCh),
					dockerClient.EXPECT().ContainerStart(gomock.Any(), "containerID", gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerLogs(gomock.Any(), "containerID", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}).Return(io.NopCloser(bytes.NewBuffer([]byte{})), nil),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(errors.New("remove error")),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: errors.New("remove error"),
		},
		{
			name: "NetworkConnect error",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)
				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(errors.New("network connection error")),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(nil),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: errors.New("network connection error"),
		},
		{
			name: "ContainerStart error",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)

				// Create channels
				waitCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerWait(gomock.Any(), "containerID", gomock.Any()).Return(waitCh, errCh),
					dockerClient.EXPECT().ContainerStart(gomock.Any(), "containerID", gomock.Any()).Return(errors.New("start container error")),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(nil),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: errors.New("start container error"),
		},
		{
			name: "ContainerWait error channel",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)

				// Create channels
				waitCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				// Write to error channel
				errCh <- errors.New("container wait error")

				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerWait(gomock.Any(), "containerID", gomock.Any()).Return(waitCh, errCh),
					dockerClient.EXPECT().ContainerStart(gomock.Any(), "containerID", gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerLogs(gomock.Any(), "containerID", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}).Return(io.NopCloser(bytes.NewBufferString("container logs")), nil),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(nil),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: errors.New("error waiting for container containerID: container wait error. container logs: container logs"),
		},
		{
			name: "Non-zero exit status",
			setupMock: func(ctrl *gomock.Controller) *mocks.MockAPIClient {
				dockerClient := mocks.NewMockAPIClient(ctrl)

				// Create channels
				waitCh := make(chan container.WaitResponse, 1)
				errCh := make(chan error, 1)

				// Write to one of the channels
				waitCh <- container.WaitResponse{StatusCode: 1}

				gomock.InOrder(
					dockerClient.EXPECT().ContainerCreate(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(container.CreateResponse{ID: "containerID"}, nil),
					dockerClient.EXPECT().NetworkConnect(gomock.Any(), "my-network", gomock.Any(), gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerWait(gomock.Any(), "containerID", gomock.Any()).Return(waitCh, errCh),
					dockerClient.EXPECT().ContainerStart(gomock.Any(), "containerID", gomock.Any()).Return(nil),
					dockerClient.EXPECT().ContainerLogs(gomock.Any(), "containerID", types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true}).Return(io.NopCloser(bytes.NewBufferString("container logs")), nil),
					dockerClient.EXPECT().ContainerRemove(gomock.Any(), "containerID", gomock.Any()).Return(nil),
				)
				return dockerClient
			},
			image:         "my-image",
			network:       "my-network",
			args:          []string{"arg1", "arg2"},
			expectedError: fmt.Errorf("unexpected exit code 1 for container containerID. container logs: container logs"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			dockerClient := tt.setupMock(ctrl)

			dockerManager := NewDockerManager(dockerClient)

			err := dockerManager.Run(tt.image, tt.network, tt.args)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
