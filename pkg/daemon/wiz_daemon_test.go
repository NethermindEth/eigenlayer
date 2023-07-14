package daemon

import (
	"encoding/json"
	"errors"
	"io"
	"path/filepath"
	"testing"

	"github.com/NethermindEth/egn/internal/common"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/data"
	mock_locker "github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/pkg/daemon/mocks"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const MockAVSLatestVersion = "v3.0.2"

func TestInit(t *testing.T) {
	// Silence logger
	log.SetOutput(io.Discard)

	tests := []struct {
		name    string
		mocker  func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager
		wantErr bool
	}{
		{
			name: "monitoring -> prev: not installed, after: installation status error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, errors.New("installation status error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: not installed, after: installed and started",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InstallStack().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: not installed, after: installation failed",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InstallStack().Return(monitoring.ErrInstallingMonitoringMngr),
					monitoringMgr.EXPECT().Cleanup(true).Return(nil),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: not installed, after: installation failed, cleanup error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InstallStack().Return(monitoring.ErrInstallingMonitoringMngr),
					monitoringMgr.EXPECT().Cleanup(true).Return(errors.New("cleanup error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: not installed, after: installation failed but no cleanup needed",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InstallStack().Return(errors.New("init error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: installed and running, after: installed and running",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Running, nil),
					monitoringMgr.EXPECT().Init().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: installed and created, after: installed and running",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Created, nil),
					monitoringMgr.EXPECT().Run().Return(nil),
					monitoringMgr.EXPECT().Init().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: installed and created, after: installed and run-error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Created, nil),
					monitoringMgr.EXPECT().Run().Return(errors.New("run error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: installed, after: installed and status error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Unknown, errors.New("status error")),
					monitoringMgr.EXPECT().Run().Return(nil),
					monitoringMgr.EXPECT().Init().Return(nil),
				)
				return monitoringMgr
			},
			wantErr: false,
		},
		{
			name: "monitoring -> prev: installed and restarting, after: installed and restarting",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Restarting, nil),
					monitoringMgr.EXPECT().Init().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: installed and broken, after: installed and re-run",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Broken, nil),
					monitoringMgr.EXPECT().Run().Return(nil),
					monitoringMgr.EXPECT().Init().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: installed and broken, after: installed and run error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Broken, nil),
					monitoringMgr.EXPECT().Run().Return(errors.New("run error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: installed and created, after: monitoring stack initialization error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Created, nil),
					monitoringMgr.EXPECT().Run().Return(nil),
					monitoringMgr.EXPECT().Init().Return(monitoring.ErrInitializingMonitoringMngr),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock monitoring manager.
			ctrl := gomock.NewController(t)

			// Create mock compose manager
			composeMgr := mocks.NewMockComposeManager(ctrl)

			// Create mock docker manager
			dockerMgr := mocks.NewMockDockerManager(ctrl)

			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)

			// Create in-memory filesystem
			afs := afero.NewMemMapFs()

			// Create DataDir
			dataDir, err := data.NewDataDir("/tmp", afs, locker)
			require.NoError(t, err)

			// Get monitoring manager mock
			monitoringMgr := tt.mocker(t, ctrl)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, composeMgr, dockerMgr, monitoringMgr, locker)
			require.NoError(t, err)

			err = daemon.Init()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestPull(t *testing.T) {
	afs := afero.NewOsFs()

	pullResult302 := PullResult{
		Version: MockAVSLatestVersion,
		Options: map[string][]Option{
			"option-returner": {
				&OptionID{
					option: option{
						name:   "main-container-name",
						target: "MAIN_SERVICE_NAME",
						help:   "Main service container name",
					},
					defValue: "option-returner",
				},
				&OptionPort{
					option: option{
						name:   "main-port",
						target: "MAIN_PORT",
						help:   "Main service server port",
					},
					defValue: 8080,
				},
				&OptionString{
					option: option{
						name:   "network-name",
						target: "NETWORK_NAME",
						help:   "Docker network name",
					},
					defValue: "eigenlayer",
					validate: true,
					Re2Regex: "^eigen.*",
				},
				&OptionInt{
					option: option{
						name:   "test-option-int",
						target: "TEST_OPTION_INT",
						help:   "Test option int",
					},
					defValue: 666,
					validate: true,
					MinValue: 0,
					MaxValue: 1000,
				},
				&OptionFloat{
					option: option{
						name:   "test-option-float",
						target: "TEST_OPTION_FLOAT",
						help:   "Test option float",
					},
					defValue: 666.666,
					validate: true,
					MinValue: 0.0,
					MaxValue: 1000.0,
				},
				&OptionBool{
					option: option{
						name:   "test-option-bool",
						target: "TEST_OPTION_BOOL",
						help:   "Test option bool",
					},
					defValue: true,
				},
				&OptionPathDir{
					option: option{
						name:   "test-option-path-dir",
						target: "TEST_OPTION_PATH_DIR",
						help:   "Test option path dir",
					},
					defValue: "/tmp",
				},
				&OptionPathFile{
					option: option{
						name:   "test-option-path-file",
						target: "TEST_OPTION_PATH_FILE",
						help:   "Test option path file",
					},
					defValue: "/tmp/test.txt",
					validate: true,
					Format:   ".txt",
				},
				&OptionURI{
					option: option{
						name:   "test-option-uri",
						target: "TEST_OPTION_URI",
						help:   "Test option uri",
					},
					defValue: "https://www.google.com",
					validate: true,
					UriScheme: []string{
						"https",
					},
				},
				&OptionSelect{
					option: option{
						name:   "test-option-enum",
						target: "TEST_OPTION_ENUM",
						help:   "Test option enum",
					},
					defValue: "option1",
					validate: true,
					Options: []string{
						"option1",
						"option2",
						"option3",
					},
				},
			},
			"health-checker": {
				&OptionID{
					option: option{
						name:   "main-container-name",
						target: "MAIN_SERVICE_NAME",
						help:   "Main service container name",
					},
					defValue: "health-checker",
				},
				&OptionPort{
					option: option{
						name:   "main-port",
						target: "MAIN_PORT",
						help:   "Main service server port",
					},
					defValue: 8090,
				},
				&OptionString{
					option: option{
						name:   "network-name",
						target: "NETWORK_NAME",
						help:   "Docker network name",
					},
					defValue: "eigenlayer",
					validate: true,
					Re2Regex: "^eigen.*",
				},
			},
		},
	}

	tests := []struct {
		name    string
		url     string
		version string
		force   bool
		want    PullResult
		mocker  func(t *testing.T, locker *mock_locker.MockLocker) *data.DataDir
		wantErr bool
	}{
		{
			name: "pull -> success",
			url:  "https://github.com/NethermindEth/mock-avs",
			want: pullResult302,
			mocker: func(t *testing.T, locker *mock_locker.MockLocker) *data.DataDir {
				tmp, err := afero.TempDir(afs, "", "egn-pull")
				require.NoError(t, err)
				dataDir, err := data.NewDataDir(tmp, afs, locker)
				require.NoError(t, err)
				return dataDir
			},
		},
		{
			name:    "pull -> success, fixed version",
			url:     "https://github.com/NethermindEth/mock-avs",
			version: MockAVSLatestVersion,
			want:    pullResult302,
			mocker: func(t *testing.T, locker *mock_locker.MockLocker) *data.DataDir {
				tmp, err := afero.TempDir(afs, "", "egn-pull")
				require.NoError(t, err)
				dataDir, err := data.NewDataDir(tmp, afs, locker)
				require.NoError(t, err)
				return dataDir
			},
		},
		{
			name:  "pull -> success, force",
			url:   "https://github.com/NethermindEth/mock-avs",
			force: true,
			want:  pullResult302,
			mocker: func(t *testing.T, locker *mock_locker.MockLocker) *data.DataDir {
				tmp, err := afero.TempDir(afs, "", "egn-pull")
				require.NoError(t, err)
				dataDir, err := data.NewDataDir(tmp, afs, locker)
				require.NoError(t, err)
				afs.MkdirAll(filepath.Join(tmp, "temp", tempID("https://github.com/NethermindEth/mock-avs")), 0o755)
				return dataDir
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock locker
			ctrl := gomock.NewController(t)
			locker := mock_locker.NewMockLocker(ctrl)

			dataDir := tt.mocker(t, locker)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, nil, nil, nil, locker)
			require.NoError(t, err)

			result, err := daemon.Pull(tt.url, tt.version, tt.force)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				// Deep check the result
				assert.Equal(t, tt.want.Version, result.Version)
				for k, profile := range tt.want.Options {
					gotProfile, ok := result.Options[k]
					require.True(t, ok)
					for _, wantOption := range profile {
						for _, gotOption := range gotProfile {
							if wantOption.Name() == gotOption.Name() {
								assert.EqualValues(t, wantOption, gotOption)
							}
						}
					}
				}
			}
		})
	}
}

func TestInstall(t *testing.T) {
	afs := afero.NewOsFs()

	tests := []struct {
		name              string
		options           InstallOptions
		monitoringTargets data.MonitoringTargets
		want              string
		mocker            func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager)
		wantErr           bool
		checkCleanup      bool
	}{
		{
			name: "install -> success, default tag",
			options: InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
			monitoringTargets: data.MonitoringTargets{
				Targets: []data.MonitoringTarget{
					{
						Service: "main-service",
						Port:    "8090",
						Path:    "/metrics",
					},
				},
			},
			want: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil),
					composeManager.EXPECT().PS(compose.DockerComposePsOptions{
						Path:   path,
						Format: "json",
						All:    true,
					}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil),
					dockerManager.EXPECT().ContainerIP("1").Return("168.128.66.1", nil),
					dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil),
					monitoringManager.EXPECT().AddTarget("http://168.128.66.1:8090", "mock-avs-default", "eigenlayer").Return(nil),
				)
			},
		},
		{
			name: "install -> success, specific tag, option-returner",
			options: InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "option-returner",
				Tag:     "specific",
			},
			monitoringTargets: data.MonitoringTargets{
				Targets: []data.MonitoringTarget{
					{
						Service: "main-service",
						Port:    "8080",
						Path:    "/metrics",
					},
				},
			},
			want: "mock-avs-specific",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-specific", "docker-compose.yml")

				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-specific", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil),
					composeManager.EXPECT().PS(compose.DockerComposePsOptions{
						Path:   path,
						Format: "json",
						All:    true,
					}).Return(`[{"ID": "2", "Service": "main-service"}]`, nil),
					dockerManager.EXPECT().ContainerIP("2").Return("168.128.66.2", nil),
					dockerManager.EXPECT().ContainerNetworks("2").Return([]string{"eigenlayer"}, nil),
					monitoringManager.EXPECT().AddTarget("http://168.128.66.2:8080", "mock-avs-specific", "eigenlayer").Return(nil),
				)
			},
		},
		{
			name: "install -> failure, bad tap version, got empty instanceID -> no install cleanup",
			options: InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "invalid-profile",
				Tag:     "default",
			},
			monitoringTargets: data.MonitoringTargets{
				Targets: []data.MonitoringTarget{
					{
						Service: "main-service",
						Port:    "8090",
						Path:    "/metrics",
					},
				},
			},
			want: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
			},
			wantErr: true,
		},
		{
			name: "install -> failure, compose create error -> install cleanup",
			options: InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
			monitoringTargets: data.MonitoringTargets{
				Targets: []data.MonitoringTarget{
					{
						Service: "main-service",
						Port:    "8090",
						Path:    "/metrics",
					},
				},
			},
			want: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(errors.New("compose create error")),
					composeManager.EXPECT().PS(compose.DockerComposePsOptions{
						Path:   path,
						Format: "json",
						All:    true,
					}).Return(`[{"ID": "3", "Service": "main-service"}]`, nil),
					dockerManager.EXPECT().ContainerIP("3").Return("168.128.66.3", nil),
					dockerManager.EXPECT().ContainerNetworks("3").Return([]string{"eigenerror"}, nil),
					monitoringManager.EXPECT().RemoveTarget("http://168.128.66.3:8090", "eigenerror").Return(nil),
				)
			},
			wantErr:      true,
			checkCleanup: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := afero.TempDir(afs, "", "egn-test-install")
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			// Create a mock compose manager
			composeManager := mocks.NewMockComposeManager(ctrl)
			// Create a mock docker manager
			dockerManager := mocks.NewMockDockerManager(ctrl)
			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Create a mock monitoring manager
			monitoringManager := mocks.NewMockMonitoringManager(ctrl)

			// Create a Datadir
			dataDir, err := data.NewDataDir(tmp, afs, locker)
			require.NoError(t, err)

			tt.mocker(tmp, composeManager, dockerManager, locker, monitoringManager)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, composeManager, dockerManager, monitoringManager, locker)
			require.NoError(t, err)

			// Pull the package
			pullResult, err := daemon.Pull(tt.options.URL, tt.options.Version, true)
			require.NoError(t, err)
			tt.options.Options = pullResult.Options[tt.options.Profile]

			// Fill option's values
			for _, option := range tt.options.Options {
				err := option.Set(option.Default())
				require.NoError(t, err)
			}

			result, err := daemon.Install(tt.options)
			if tt.wantErr {
				require.Error(t, err)
				// Check if temp dir was removed
				tID := tempID(tt.options.URL)
				exists, err := afero.DirExists(afs, filepath.Join(tmp, "temp", tID))
				require.NoError(t, err)
				assert.False(t, exists)

				// Check if instance dir was removed
				if tt.checkCleanup {
					exists, err = afero.DirExists(afs, filepath.Join(tmp, "nodes", tt.want))
					require.NoError(t, err)
					assert.False(t, exists)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, result)

				// Check the instance was installed
				exists, err := afero.DirExists(afs, filepath.Join(tmp, "nodes", tt.want))
				require.NoError(t, err)
				assert.True(t, exists)

				files := []string{".lock", "docker-compose.yml", ".env", "state.json"}
				for _, file := range files {
					exists, err = afero.Exists(afs, filepath.Join(tmp, "nodes", tt.want, file))
					assert.NoError(t, err)
					assert.True(t, exists)
				}

				// Validate state.json
				var instance data.Instance
				stateData, err := afero.ReadFile(afs, filepath.Join(tmp, "nodes", tt.want, "state.json"))
				require.NoError(t, err)
				err = json.Unmarshal(stateData, &instance)
				require.NoError(t, err)

				assert.Equal(t, "mock-avs", instance.Name)
				assert.Equal(t, tt.options.URL, instance.URL)
				assert.Equal(t, tt.options.Version, instance.Version)
				assert.Equal(t, tt.options.Profile, instance.Profile)
				assert.Equal(t, tt.options.Tag, instance.Tag)
				assert.Equal(t, tt.monitoringTargets, instance.MonitoringTargets)
			}
		})
	}
}

func TestRun(t *testing.T) {
	afs := afero.NewOsFs()

	tests := []struct {
		name       string
		instanceID string
		mocker     func(string, *mocks.MockComposeManager, *mocks.MockDockerManager, *mock_locker.MockLocker, *mocks.MockMonitoringManager)
		options    *InstallOptions
		wantErr    bool
	}{
		{
			name:       "success",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				// Init, install and run
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil).Times(2)
				composeManager.EXPECT().PS(compose.DockerComposePsOptions{
					Path:   path,
					Format: "json",
					All:    true,
				}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil).Times(2)
				dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil).Times(2)
				dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil).Times(2)
				monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil).Times(2)
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
		},
		{
			name:       "failure, not installed instance",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
			},
			wantErr: true,
		},
		{
			name:       "failure, Up error",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				// Init, install and run
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil)
				composeManager.EXPECT().PS(compose.DockerComposePsOptions{
					Path:   path,
					Format: "json",
					All:    true,
				}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil).Times(2)
				dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil).Times(2)
				dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil).Times(2)
				monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil).Times(2)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(errors.New("error"))
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := afero.TempDir(afs, "", "egn-test-run")
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			// Create a mock compose manager
			composeManager := mocks.NewMockComposeManager(ctrl)
			// Create a mock docker manager
			dockerManager := mocks.NewMockDockerManager(ctrl)
			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Create a mock monitoring manager
			monitoringManager := mocks.NewMockMonitoringManager(ctrl)

			// Create a Datadir
			dataDir, err := data.NewDataDir(tmp, afs, locker)
			require.NoError(t, err)

			tt.mocker(tmp, composeManager, dockerManager, locker, monitoringManager)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, composeManager, dockerManager, monitoringManager, locker)
			require.NoError(t, err)

			if tt.options != nil {
				// Pull the package
				pullResult, err := daemon.Pull(tt.options.URL, tt.options.Version, true)
				require.NoError(t, err)
				tt.options.Options = pullResult.Options[tt.options.Profile]

				// Fill option's values
				for _, option := range tt.options.Options {
					err := option.Set(option.Default())
					require.NoError(t, err)
				}

				_, err = daemon.Install(*tt.options)
				require.NoError(t, err)
			}

			err = daemon.Run(tt.instanceID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestStop(t *testing.T) {
	afs := afero.NewOsFs()

	tests := []struct {
		name       string
		instanceID string
		mocker     func(string, *mocks.MockComposeManager, *mocks.MockDockerManager, *mock_locker.MockLocker, *mocks.MockMonitoringManager)
		options    *InstallOptions
		wantErr    bool
	}{
		{
			name:       "success",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				gomock.InOrder(
					// Init and install
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil),
					composeManager.EXPECT().PS(compose.DockerComposePsOptions{
						Path:   path,
						Format: "json",
						All:    true,
					}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil),
					dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil),
					dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil),
					monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil),
					// Stop
					composeManager.EXPECT().Stop(compose.DockerComposeStopOptions{Path: path}).Return(nil),
				)
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
		},
		{
			name:       "failure, not installed instance",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
			},
			wantErr: true,
		},
		{
			name:       "failure, Stop error",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				gomock.InOrder(
					// Init and install
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
					composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil),
					composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil),
					composeManager.EXPECT().PS(compose.DockerComposePsOptions{
						Path:   path,
						Format: "json",
						All:    true,
					}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil),
					dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil),
					dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil),
					monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil),
					// Stop
					composeManager.EXPECT().Stop(compose.DockerComposeStopOptions{Path: path}).Return(errors.New("error")),
				)
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := afero.TempDir(afs, "", "egn-test-stop")
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			// Create a mock compose manager
			composeManager := mocks.NewMockComposeManager(ctrl)
			// Create a mock docker manager
			dockerManager := mocks.NewMockDockerManager(ctrl)
			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Create a mock monitoring manager
			monitoringManager := mocks.NewMockMonitoringManager(ctrl)

			// Create a Datadir
			dataDir, err := data.NewDataDir(tmp, afs, locker)
			require.NoError(t, err)

			tt.mocker(tmp, composeManager, dockerManager, locker, monitoringManager)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, composeManager, dockerManager, monitoringManager, locker)
			require.NoError(t, err)

			if tt.options != nil {
				// Pull the package
				pullResult, err := daemon.Pull(tt.options.URL, tt.options.Version, true)
				require.NoError(t, err)
				tt.options.Options = pullResult.Options[tt.options.Profile]

				// Fill option's values
				for _, option := range tt.options.Options {
					err := option.Set(option.Default())
					require.NoError(t, err)
				}

				_, err = daemon.Install(*tt.options)
				require.NoError(t, err)
			}

			err = daemon.Stop(tt.instanceID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestUninstall(t *testing.T) {
	afs := afero.NewOsFs()

	tests := []struct {
		name       string
		instanceID string
		mocker     func(string, *mocks.MockComposeManager, *mocks.MockDockerManager, *mock_locker.MockLocker, *mocks.MockMonitoringManager)
		options    *InstallOptions
		wantErr    bool
	}{
		{
			name:       "success",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				// Init and install
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil)
				composeManager.EXPECT().PS(compose.DockerComposePsOptions{
					Path:   path,
					Format: "json",
					All:    true,
				}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil).Times(2)
				dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil).Times(2)
				dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil).Times(2)
				monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil)
				// Uninstall
				monitoringManager.EXPECT().RemoveTarget("http://168.66.44.1:8090", "eigenlayer").Return(nil)
				composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: path}).Return(nil)
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
		},
		{
			name:       "failure, not installed instance",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
			},
			wantErr: true,
		},
		{
			name:       "failure, Down error",
			instanceID: "mock-avs-default",
			mocker: func(tmp string, composeManager *mocks.MockComposeManager, dockerManager *mocks.MockDockerManager, locker *mock_locker.MockLocker, monitoringManager *mocks.MockMonitoringManager) {
				path := filepath.Join(tmp, "nodes", "mock-avs-default", "docker-compose.yml")

				// Init and install
				gomock.InOrder(
					locker.EXPECT().New(filepath.Join(tmp, "nodes", "mock-avs-default", ".lock")).Return(locker),
					locker.EXPECT().Lock().Return(nil),
					locker.EXPECT().Locked().Return(true),
					locker.EXPECT().Unlock().Return(nil),
				)
				composeManager.EXPECT().Create(compose.DockerComposeCreateOptions{Path: path, Build: true}).Return(nil)
				composeManager.EXPECT().Up(compose.DockerComposeUpOptions{Path: path}).Return(nil)
				composeManager.EXPECT().PS(compose.DockerComposePsOptions{
					Path:   path,
					Format: "json",
					All:    true,
				}).Return(`[{"ID": "1", "Service": "main-service"}]`, nil).Times(2)
				dockerManager.EXPECT().ContainerIP("1").Return("168.66.44.1", nil).Times(2)
				dockerManager.EXPECT().ContainerNetworks("1").Return([]string{"eigenlayer"}, nil).Times(2)
				monitoringManager.EXPECT().AddTarget("http://168.66.44.1:8090", "mock-avs-default", "eigenlayer").Return(nil)
				// Uninstall
				monitoringManager.EXPECT().RemoveTarget("http://168.66.44.1:8090", "eigenlayer").Return(nil)
				composeManager.EXPECT().Down(compose.DockerComposeDownOptions{Path: path}).Return(errors.New("error"))
			},
			options: &InstallOptions{
				URL:     "https://github.com/NethermindEth/mock-avs",
				Version: MockAVSLatestVersion,
				Profile: "health-checker",
				Tag:     "default",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, err := afero.TempDir(afs, "", "egn-test-uninstall")
			require.NoError(t, err)

			ctrl := gomock.NewController(t)
			// Create a mock compose manager
			composeManager := mocks.NewMockComposeManager(ctrl)
			// Create a mock docker manager
			dockerManager := mocks.NewMockDockerManager(ctrl)
			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)
			// Create a mock monitoring manager
			monitoringManager := mocks.NewMockMonitoringManager(ctrl)

			// Create a Datadir
			dataDir, err := data.NewDataDir(tmp, afs, locker)
			require.NoError(t, err)

			tt.mocker(tmp, composeManager, dockerManager, locker, monitoringManager)

			// Create a daemon
			daemon, err := NewWizDaemon(dataDir, composeManager, dockerManager, monitoringManager, locker)
			require.NoError(t, err)

			if tt.options != nil {
				// Pull the package
				pullResult, err := daemon.Pull(tt.options.URL, tt.options.Version, true)
				require.NoError(t, err)
				tt.options.Options = pullResult.Options[tt.options.Profile]

				// Fill option's values
				for _, option := range tt.options.Options {
					err := option.Set(option.Default())
					require.NoError(t, err)
				}

				_, err = daemon.Install(*tt.options)
				require.NoError(t, err)
			}

			err = daemon.Uninstall(tt.instanceID)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// Check the instance was uninstalled
				exists, err := afero.DirExists(afs, filepath.Join(tmp, "nodes", tt.instanceID))
				require.NoError(t, err)
				assert.False(t, exists)
			}
		})
	}
}
