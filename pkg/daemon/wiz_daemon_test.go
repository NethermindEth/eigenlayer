package daemon

import (
	"errors"
	"io"
	"testing"

	"github.com/NethermindEth/egn/internal/common"
	mock_locker "github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/pkg/daemon/mocks"
	"github.com/golang/mock/gomock"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

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
			fs := afero.NewMemMapFs()

			// Get monitoring manager mock
			monitoringMgr := tt.mocker(t, ctrl)

			// Create a daemon
			daemon, err := NewWizDaemon(composeMgr, dockerMgr, monitoringMgr, fs, locker)
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
