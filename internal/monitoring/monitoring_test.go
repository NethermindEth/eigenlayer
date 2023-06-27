package monitoring

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/data"
	mock_locker "github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/NethermindEth/egn/internal/monitoring/mocks"
	"github.com/NethermindEth/egn/internal/monitoring/services/types"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitStack(t *testing.T) {
	// Silence logger
	log.SetOutput(io.Discard)

	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		require.NoError(t, err)
		userDataHome = filepath.Join(userHome, ".local", "share")
	}

	okLocker := func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
		// Create a mock locker
		locker := mock_locker.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		gomock.InOrder(
			locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		// stack.Installed() lock
		gomock.InOrder(
			locker.EXPECT().Lock().Return(nil),
			locker.EXPECT().Locked().Return(true),
			locker.EXPECT().Unlock().Return(nil),
		)
		return locker
	}
	onlyNewLocker := func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
		// Create a mock locker
		locker := mock_locker.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker)
		return locker
	}

	// Setup mock http server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Listen for the POST request to /-/reload
		if r.URL.Path == "/-/reload" && r.Method == http.MethodPost {
			// All good
			w.WriteHeader(http.StatusOK)
		} else if r.Method != http.MethodGet {
			// Unexpected method
			w.WriteHeader(http.StatusMethodNotAllowed)
		} else {
			// Unexpected path
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()
	serverPort := strings.Split(server.URL, ":")[2]

	tests := []struct {
		name         string
		mockerLocker func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker
		mocker       func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager)
		dotenv       map[string]string
		wantErr      bool
	}{
		{
			name:         "ok, 1 service, port not occupied",
			mockerLocker: okLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					servicer.EXPECT().Setup(dotenv).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
		},
		{
			name:         "ok, 2 services",
			mockerLocker: okLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE1_PORT": "9000",
					"NODE2_PORT": "9001",
				}
				service1, service2 := mocks.NewMockServiceAPI(ctrl), mocks.NewMockServiceAPI(ctrl)

				// Expect the service to be triggered
				gomock.InOrder(
					service1.EXPECT().DotEnv().Return(map[string]string{
						"NODE1_PORT": "9000",
					}),
					service1.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					service1.EXPECT().Setup(dotenv).Return(nil),
				)
				gomock.InOrder(
					service2.EXPECT().DotEnv().Return(map[string]string{
						"NODE2_PORT": "9000",
					}),
					service2.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					service2.EXPECT().Setup(dotenv).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)

				return []ServiceAPI{
					service1,
					service2,
				}, composeManager
			},
		},
		{
			name:         "ok, 1 service, port occupied",
			mockerLocker: okLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				// Convert serverPort to int
				p, err := strconv.Atoi(serverPort)
				require.NoError(t, err)
				sp := strconv.Itoa(p + 1)

				dotenv := map[string]string{
					"NODE_PORT": sp,
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(map[string]string{"NODE_PORT": serverPort}),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					servicer.EXPECT().Setup(dotenv).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
		},
		{
			name:         "error, 1 service, port not int",
			mockerLocker: onlyNewLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "3RR0R",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				servicer.EXPECT().DotEnv().Return(dotenv)

				composeManager := mocks.NewMockComposeManager(ctrl)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name:         "error, 1 service, invalid port",
			mockerLocker: onlyNewLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "0",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				servicer.EXPECT().DotEnv().Return(dotenv)

				composeManager := mocks.NewMockComposeManager(ctrl)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name:         "error, 1 service, init service error",
			mockerLocker: onlyNewLocker,
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(errors.New("error")),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name: "error, 1 service, stack setup error",
			mockerLocker: func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(errors.New("error")),
				)
				return locker
			},
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name: "error, 1 service, service setup error",
			mockerLocker: func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				return locker
			},
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					servicer.EXPECT().Setup(dotenv).Return(errors.New("error")),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name: "error, 1 service, create error",
			mockerLocker: func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				return locker
			},
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					servicer.EXPECT().Setup(dotenv).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(errors.New("error"))

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
		{
			name: "error, 1 service, run error",
			mockerLocker: func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)

				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				return locker
			},
			mocker: func(t *testing.T, ctrl *gomock.Controller, stack *data.MonitoringStack) ([]ServiceAPI, *mocks.MockComposeManager) {
				dotenv := map[string]string{
					"NODE_PORT": "9000",
				}
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				gomock.InOrder(
					servicer.EXPECT().DotEnv().Return(dotenv),
					servicer.EXPECT().Init(types.ServiceOptions{
						Stack:  stack,
						Dotenv: dotenv,
					}).Return(nil),
					servicer.EXPECT().Setup(dotenv).Return(nil),
				)

				composeManager := mocks.NewMockComposeManager(ctrl)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: filepath.Join(stack.Path(), "docker-compose.yml")}).Return(errors.New("error"))

				return []ServiceAPI{
					servicer,
				}, composeManager
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Helper()
			// Create a mock controller
			ctrl := gomock.NewController(t)
			locker := tt.mockerLocker(t, ctrl)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				[]ServiceAPI{},
				mocks.NewMockComposeManager(ctrl),
				mocks.NewMockDockerManager(ctrl),
				afero.NewMemMapFs(),
				locker,
			)

			services, composeManager := tt.mocker(t, ctrl, manager.stack)
			manager.services = services
			manager.composeManager = composeManager

			// Init the stack
			err := manager.InitStack()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Check the stack is installed
				installed, err := manager.stack.Installed()
				assert.NoError(t, err)
				assert.True(t, installed)
			}
		})
	}
}

func TestAddAndRemoveTarget(t *testing.T) {
	okLocker := func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker {
		// Create a mock locker
		locker := mock_locker.NewMockLocker(ctrl)

		// Expect the lock to be acquired
		userDataHome := os.Getenv("XDG_DATA_HOME")
		if userDataHome == "" {
			userHome, err := os.UserHomeDir()
			require.NoError(t, err)
			userDataHome = filepath.Join(userHome, ".local", "share")
		}
		locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker)

		return locker
	}

	tests := []struct {
		name         string
		mockerLocker func(t *testing.T, ctrl *gomock.Controller) *mock_locker.MockLocker
		services     func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI
		target       string
		add          bool
		wantErr      bool
	}{
		{
			name:         "add, ok, 1 service",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				servicer.EXPECT().AddTarget(target).Return(nil)
				return []ServiceAPI{
					servicer,
				}
			},
			target: "localhost:9000",
			add:    true,
		},
		{
			name:         "add, ok, 2 services",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				service1, service2 := mocks.NewMockServiceAPI(ctrl), mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				service1.EXPECT().AddTarget(target).Return(nil)
				service2.EXPECT().AddTarget(target).Return(nil)
				return []ServiceAPI{
					service1,
					service2,
				}
			},
			target: "http://localhost:9000",
			add:    true,
		},
		{
			name:         "add, ok, 2 services, 1 error",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				service1, service2 := mocks.NewMockServiceAPI(ctrl), mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				service1.EXPECT().AddTarget(target).Return(nil)
				service2.EXPECT().AddTarget(target).Return(errors.New("error"))
				return []ServiceAPI{
					service1,
					service2,
				}
			},
			target:  "http://localhost:9000",
			wantErr: true,
			add:     true,
		},
		{
			name:         "remove, ok, 1 service",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				servicer := mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				servicer.EXPECT().RemoveTarget(target).Return(nil)
				return []ServiceAPI{
					servicer,
				}
			},
			target: "localhost:9000",
		},
		{
			name:         "remove, ok, 2 services",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				service1, service2 := mocks.NewMockServiceAPI(ctrl), mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				service1.EXPECT().RemoveTarget(target).Return(nil)
				service2.EXPECT().RemoveTarget(target).Return(nil)
				return []ServiceAPI{
					service1,
					service2,
				}
			},
			target: "http://localhost:9000",
		},
		{
			name:         "remove, ok, 2 services, 1 error",
			mockerLocker: okLocker,
			services: func(t *testing.T, ctrl *gomock.Controller, target string) []ServiceAPI {
				service1, service2 := mocks.NewMockServiceAPI(ctrl), mocks.NewMockServiceAPI(ctrl)
				// Expect the service to be triggered
				service1.EXPECT().RemoveTarget(target).Return(nil)
				service2.EXPECT().RemoveTarget(target).Return(errors.New("error"))
				return []ServiceAPI{
					service1,
					service2,
				}
			},
			target:  "http://localhost:9000",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock controller
			ctrl := gomock.NewController(t)
			locker := tt.mockerLocker(t, ctrl)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				tt.services(t, ctrl, tt.target),
				mocks.NewMockComposeManager(ctrl),
				mocks.NewMockDockerManager(ctrl),
				afero.NewMemMapFs(),
				locker,
			)

			var err error
			if tt.add {
				// Add the target
				err = manager.AddTarget(tt.target)
			} else {
				// Remove the target
				err = manager.RemoveTarget(tt.target)
			}
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestRun(t *testing.T) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		require.NoError(t, err)
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	composePath := filepath.Join(userDataHome, ".eigen", "monitoring", "docker-compose.yml")

	tests := []struct {
		name    string
		mocker  func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager
		wantErr bool
	}{
		{
			name: "ok",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager {
				composeManager := mocks.NewMockComposeManager(ctrl)
				// Expect the compose manager to be triggered
				gomock.InOrder(
					composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: composePath}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: composePath}).Return(nil),
				)
				return composeManager
			},
		},
		{
			name: "down error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager {
				composeManager := mocks.NewMockComposeManager(ctrl)
				// Expect the compose manager to be triggered
				composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: composePath}).Return(errors.New("error"))
				return composeManager
			},
			wantErr: true,
		},
		{
			name: "up error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager {
				composeManager := mocks.NewMockComposeManager(ctrl)
				// Expect the compose manager to be triggered
				gomock.InOrder(
					composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: composePath}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: composePath}).Return(errors.New("error")),
				)
				return composeManager
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create an in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create a mock controller
			ctrl := gomock.NewController(t)

			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Expect the lock to be acquired
			locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				[]ServiceAPI{mocks.NewMockServiceAPI(ctrl)},
				tt.mocker(t, ctrl),
				mocks.NewMockDockerManager(ctrl),
				afs,
				locker,
			)

			// Run the stack
			err := manager.Run()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStop(t *testing.T) {
	// Silence logger
	log.SetOutput(io.Discard)

	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		require.NoError(t, err)
		userDataHome = filepath.Join(userHome, ".local", "share")
	}
	composePath := filepath.Join(userDataHome, ".eigen", "monitoring", "docker-compose.yml")

	tests := []struct {
		name    string
		mocker  func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager
		wantErr bool
	}{
		{
			name: "ok",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager {
				composeManager := mocks.NewMockComposeManager(ctrl)
				// Expect the compose manager to be triggered
				gomock.InOrder(
					composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: composePath}).Return(nil),
				)
				return composeManager
			},
		},
		{
			name: "down error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockComposeManager {
				composeManager := mocks.NewMockComposeManager(ctrl)
				// Expect the compose manager to be triggered
				composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: composePath}).Return(errors.New("error"))
				return composeManager
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock controller
			ctrl := gomock.NewController(t)

			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Expect the lock to be acquired
			locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				[]ServiceAPI{mocks.NewMockServiceAPI(ctrl)},
				tt.mocker(t, ctrl),
				mocks.NewMockDockerManager(ctrl),
				afero.NewMemMapFs(),
				locker,
			)

			// Run the stack
			err := manager.Stop()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestStatus(t *testing.T) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		require.NoError(t, err)
		userDataHome = filepath.Join(userHome, ".local", "share")
	}

	tests := []struct {
		name    string
		mocker  func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager
		want    common.Status
		wantErr bool
	}{
		{
			name: "ok",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				gomock.InOrder(
					dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Running, nil),
					dockerManager.EXPECT().ContainerStatus(PrometheusContainerName).Return(common.Running, nil),
					dockerManager.EXPECT().ContainerStatus(NodeExporterContainerName).Return(common.Running, nil),
				)
				return dockerManager
			},
			want: common.Running,
		},
		{
			name: "error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Unknown, errors.New("error"))
				return dockerManager
			},
			want:    common.Unknown,
			wantErr: true,
		},
		{
			name: "Restarting",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				gomock.InOrder(
					dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Restarting, nil),
					dockerManager.EXPECT().ContainerStatus(PrometheusContainerName).Return(common.Restarting, nil),
					dockerManager.EXPECT().ContainerStatus(NodeExporterContainerName).Return(common.Restarting, nil),
				)
				return dockerManager
			},
			want: common.Restarting,
		},
		{
			name: "Paused",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Paused, nil)
				return dockerManager
			},
			want:    common.Broken,
			wantErr: true,
		},
		{
			name: "Exited",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Exited, nil)
				return dockerManager
			},
			want:    common.Broken,
			wantErr: true,
		},
		{
			name: "Dead",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockDockerManager {
				dockerManager := mocks.NewMockDockerManager(ctrl)
				// Expect the docker manager to be triggered
				dockerManager.EXPECT().ContainerStatus(GrafanaContainerName).Return(common.Dead, nil)
				return dockerManager
			},
			want:    common.Broken,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock controller
			ctrl := gomock.NewController(t)

			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Expect the lock to be acquired
			locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				[]ServiceAPI{mocks.NewMockServiceAPI(ctrl)},
				mocks.NewMockComposeManager(ctrl),
				tt.mocker(t, ctrl),
				afero.NewMemMapFs(),
				locker,
			)

			// Run the stack
			status, err := manager.Status()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, status)
		})
	}
}

func TestInstallationStatus(t *testing.T) {
	userDataHome := os.Getenv("XDG_DATA_HOME")
	if userDataHome == "" {
		userHome, err := os.UserHomeDir()
		require.NoError(t, err)
		userDataHome = filepath.Join(userHome, ".local", "share")
	}

	tests := []struct {
		name    string
		mocker  func(t *testing.T, ctrl *gomock.Controller) (afero.Fs, *mock_locker.MockLocker)
		want    common.Status
		wantErr bool
	}{
		{
			name: "installed",
			mocker: func(t *testing.T, ctrl *gomock.Controller) (afero.Fs, *mock_locker.MockLocker) {
				fs := afero.NewMemMapFs()
				// Recreate installed monitoring
				fs.MkdirAll(filepath.Join(userDataHome, ".eigen", "monitoring"), 0o755)
				fs.Create(filepath.Join(userDataHome, ".eigen", "monitoring", "docker-compose.yml"))
				fs.Create(filepath.Join(userDataHome, ".eigen", "monitoring", ".env"))

				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)
				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)

				return fs, locker
			},
			want: common.Installed,
		},
		{
			name: "not installed",
			mocker: func(t *testing.T, ctrl *gomock.Controller) (afero.Fs, *mock_locker.MockLocker) {
				fs := afero.NewMemMapFs()

				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)
				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)

				return fs, locker
			},
			want: common.NotInstalled,
		},
		{
			name: "lock error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) (afero.Fs, *mock_locker.MockLocker) {
				fs := afero.NewMemMapFs()

				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)
				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(errors.New("error")),
				)

				return fs, locker
			},
			want:    common.Unknown,
			wantErr: true,
		},
		{
			name: "unlock error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) (afero.Fs, *mock_locker.MockLocker) {
				fs := afero.NewMemMapFs()

				// Create a mock locker
				locker := mock_locker.NewMockLocker(ctrl)
				// Expect the lock to be acquired
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(userDataHome, ".eigen", "monitoring", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(errors.New("error")),
				)

				return fs, locker
			},
			want:    common.Unknown,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock controller
			ctrl := gomock.NewController(t)

			fs, locker := tt.mocker(t, ctrl)

			// Create a monitoring manager
			manager := NewMonitoringManager(
				[]ServiceAPI{mocks.NewMockServiceAPI(ctrl)},
				mocks.NewMockComposeManager(ctrl),
				mocks.NewMockDockerManager(ctrl),
				fs,
				locker,
			)

			// Run the stack
			status, err := manager.InstallationStatus()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.want, status)
		})
	}
}
