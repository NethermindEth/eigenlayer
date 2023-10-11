package cli

import (
	"fmt"
	"testing"
	"time"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	prompterMock "github.com/NethermindEth/eigenlayer/cli/prompter/mocks"
	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	instanceId := "mock-avs-default"
	tc := []struct {
		name   string
		args   []string
		mocker func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter)
		err    error
	}{
		{
			name: "update to latest version",
			args: []string{instanceId},
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return(instanceId, nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run(instanceId),
				)
			},
		},
		{
			name: "update to fixed version",
			args: []string{instanceId, common.MockAvsPkg.Version()},
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{
						Version: common.MockAvsPkg.Version(),
					}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return(instanceId, nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run(instanceId),
				)
			},
		},
		{
			name: "update to fixed commit",
			args: []string{instanceId, common.MockAvsPkg.CommitHash()},
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{
						Commit: common.MockAvsPkg.CommitHash(),
					}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return(instanceId, nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run(instanceId),
				)
			},
		},
		{
			name: "update with backup",
			args: []string{instanceId, "--backup"},
			err:  nil,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return(fmt.Sprintf("%s-%d", instanceId, time.Now().Unix()), nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return(instanceId, nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run(instanceId),
				)
			},
		},
		{
			name: "update backup error",
			args: []string{instanceId, "--backup"},
			err:  assert.AnError,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return("", assert.AnError),
				)
			},
		},
		{
			name: "update with backup, fails to uninstall current instance, restore backup successfully",
			args: []string{instanceId, "--backup"},
			err:  nil,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return(fmt.Sprintf("%s-%d", instanceId, time.Now().Unix()), nil),
					d.EXPECT().Uninstall(instanceId).Return(assert.AnError),
					d.EXPECT().Restore(gomock.Any(), false).Return(nil),
				)
			},
		},
		{
			name: "update with backup, fails to uninstall current instance, restore backup fails",
			args: []string{instanceId, "--backup"},
			err:  assert.AnError,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return(fmt.Sprintf("%s-%d", instanceId, time.Now().Unix()), nil),
					d.EXPECT().Uninstall(instanceId).Return(assert.AnError),
					d.EXPECT().Restore(gomock.Any(), false).Return(assert.AnError),
				)
			},
		},
		{
			name: "update with backup, fails to install new instance, restore backup successfully",
			args: []string{instanceId, "--backup"},
			err:  nil,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return(fmt.Sprintf("%s-%d", instanceId, time.Now().Unix()), nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return("", assert.AnError),
					d.EXPECT().Restore(gomock.Any(), false).Return(nil),
				)
			},
		},
		{
			name: "update with backup, fails to install new instance, restore backup fails",
			args: []string{instanceId, "--backup"},
			err:  assert.AnError,
			mocker: func(ctrl *gomock.Controller, d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				oldOption := daemonMock.NewMockOption(ctrl)
				newOption := daemonMock.NewMockOption(ctrl)
				mergedOption := daemonMock.NewMockOption(ctrl)

				oldOption.EXPECT().IsSet().Return(true)
				oldOption.EXPECT().Value().Return("old-value", nil)
				oldOption.EXPECT().Name().Return("old-option").Times(3)
				mergedOption.EXPECT().IsSet().Return(true).Times(2)
				mergedOption.EXPECT().Value().Return("old-value", nil).Times(2)
				mergedOption.EXPECT().Name().Return("old-option").Times(2)
				mergedOption.EXPECT().Help().Return("option help")

				gomock.InOrder(
					d.EXPECT().PullUpdate(instanceId, daemon.PullTarget{}).Return(daemon.PullUpdateResult{
						Name:          "mock-avs",
						Tag:           "default",
						Url:           common.MockAvsPkg.Repo(),
						Profile:       "option-returner",
						OldVersion:    "v5.4.0",
						NewVersion:    common.MockAvsPkg.Version(),
						OldCommit:     "b64c50c15e53ae7afebbdbe210b834d1ee471043",
						NewCommit:     common.MockAvsPkg.CommitHash(),
						HasPlugin:     true,
						OldOptions:    []daemon.Option{oldOption},
						NewOptions:    []daemon.Option{newOption},
						MergedOptions: []daemon.Option{mergedOption},
						HardwareRequirements: daemon.HardwareRequirements{
							MinCPUCores:                 2,
							MinRAM:                      2048,
							MinFreeSpace:                5120,
							StopIfRequirementsAreNotMet: true,
						},
					}, nil),
					d.EXPECT().Backup(instanceId).Return(fmt.Sprintf("%s-%d", instanceId, time.Now().Unix()), nil),
					d.EXPECT().Uninstall(instanceId).Return(nil),
					d.EXPECT().Install(daemon.InstallOptions{
						Name:    "mock-avs",
						Tag:     "default",
						URL:     common.MockAvsPkg.Repo(),
						Profile: "option-returner",
						Version: common.MockAvsPkg.Version(),
						Commit:  common.MockAvsPkg.CommitHash(),
						Options: []daemon.Option{mergedOption},
					}).Return("", assert.AnError),
					d.EXPECT().Restore(gomock.Any(), false).Return(assert.AnError),
				)
			},
		},
		{
			name: "invalid arguments, instance id is required",
			args: []string{},
			err:  fmt.Errorf("%w: instance-id is required", ErrInvalidNumberOfArgs),
		},
		{
			name: "invalid arguments, more than 2 arguments",
			args: []string{"instance-id", "version", "extra-arg"},
			err:  fmt.Errorf("%w: too many arguments", ErrInvalidNumberOfArgs),
		},
		{
			name: "invalid arguments, invalid version or commit",
			args: []string{"instance-id", "invalid-version"},
			err:  fmt.Errorf("%w: invalid version or commit", ErrInvalidArgs),
		},
	}
	for _, tt := range tc {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			d := daemonMock.NewMockDaemon(ctrl)
			p := prompterMock.NewMockPrompter(ctrl)
			if tt.mocker != nil {
				tt.mocker(ctrl, d, p)
			}

			updateCmd := UpdateCmd(d, p)

			updateCmd.SetArgs(tt.args)
			err := updateCmd.Execute()

			if tt.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tt.err.Error())
			}
		})
	}
}
