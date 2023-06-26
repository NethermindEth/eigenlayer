package daemon

import (
	"errors"
	"testing"

	"github.com/NethermindEth/egn/internal/common"
	mock_locker "github.com/NethermindEth/egn/internal/locker/mocks"
	"github.com/NethermindEth/egn/pkg/daemon/mocks"
	"github.com/golang/mock/gomock"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestInit(t *testing.T) {
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
					monitoringMgr.EXPECT().InitStack().Return(nil),
					monitoringMgr.EXPECT().Status().Return(common.Running, nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: not installed, after: installed and broken",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InitStack().Return(nil),
					monitoringMgr.EXPECT().Status().Return(common.Broken, nil),
					monitoringMgr.EXPECT().Run().Return(nil),
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: not installed, after: installed and status error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InitStack().Return(nil),
					monitoringMgr.EXPECT().Status().Return(common.Unknown, errors.New("status error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: not installed, after: installation failed",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.NotInstalled, nil),
					monitoringMgr.EXPECT().InitStack().Return(errors.New("installation failed")),
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
				)
				return monitoringMgr
			},
		},
		{
			name: "monitoring -> prev: installed, after: installed and status error",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Unknown, errors.New("status error")),
				)
				return monitoringMgr
			},
			wantErr: true,
		},
		{
			name: "monitoring -> prev: installed and restarting, after: installed and restarting",
			mocker: func(t *testing.T, ctrl *gomock.Controller) *mocks.MockMonitoringManager {
				monitoringMgr := mocks.NewMockMonitoringManager(ctrl)
				gomock.InOrder(
					monitoringMgr.EXPECT().InstallationStatus().Return(common.Installed, nil),
					monitoringMgr.EXPECT().Status().Return(common.Restarting, nil),
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock monitoring manager.
			ctrl := gomock.NewController(t)

			// Create mock compose manager
			composeMgr := mocks.NewMockComposeManager(ctrl)

			// Create a mock locker
			locker := mock_locker.NewMockLocker(ctrl)

			// Create in-memory filesystem
			fs := afero.NewMemMapFs()

			// Get monitoring manager mock
			monitoringMgr := tt.mocker(t, ctrl)

			// Create a daemon
			daemon, err := NewWizDaemon(composeMgr, monitoringMgr, fs, locker)
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
