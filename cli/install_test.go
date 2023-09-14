package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	daemonMock "github.com/NethermindEth/eigenlayer/cli/mocks"
	prompterMock "github.com/NethermindEth/eigenlayer/cli/prompter/mocks"
	"github.com/NethermindEth/eigenlayer/internal/common"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
)

func TestInstall(t *testing.T) {
	ts := []struct {
		name       string
		args       []string
		err        error
		daemonMock func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter)
	}{
		{
			name: "no arguments",
			args: []string{},
			err:  fmt.Errorf("%w: accepts 1 arg, received 0", ErrInvalidNumberOfArgs),
		},
		{
			name: "more than one argument",
			args: []string{"arg1", "arg2"},
			err:  fmt.Errorf("%w: accepts 1 arg, received 2", ErrInvalidNumberOfArgs),
		},
		{
			name: "invalid URL",
			args: []string{"invalid-url"},
			err:  fmt.Errorf("%w: parse \"invalid-url\": invalid URI for request", ErrInvalidURL),
		},
		{
			name: "valid arguments, run confirmed",
			args: []string{common.MockAvsPkg.Repo()},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {
									MinCPUCores:                 2,
									MinRAM:                      2048,
									MinFreeSpace:                5120,
									StopIfRequirementsAreNotMet: true,
								},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{
						MinCPUCores:                 2,
						MinRAM:                      2048,
						MinFreeSpace:                5120,
						StopIfRequirementsAreNotMet: true,
					}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run("mock-avs-pkg-default").Return(nil),
				)
			},
		},
		{
			name: "valid arguments, run confirmed, init monitoring error",
			args: []string{common.MockAvsPkg.Repo()},
			err:  assert.AnError,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(assert.AnError),
				)
			},
		},
		{
			name: "valid arguments, run confirmed and failed",
			args: []string{common.MockAvsPkg.Repo()},
			err:  assert.AnError,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run("mock-avs-pkg-default").Return(assert.AnError),
				)
			},
		},
		{
			name: "valid arguments, run confirm error",
			args: []string{common.MockAvsPkg.Repo()},
			err:  assert.AnError,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, assert.AnError),
				)
			},
		},
		{
			name: "valid arguments, with --yes",
			args: []string{common.MockAvsPkg.Repo(), "--yes"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					d.EXPECT().Run("mock-avs-pkg-default").Return(nil),
				)
			},
		},
		{
			name: "valid arguments, with --yes, run error",
			args: []string{common.MockAvsPkg.Repo(), "--yes"},
			err:  assert.AnError,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					d.EXPECT().Run("mock-avs-pkg-default").Return(assert.AnError),
				)
			},
		},
		{
			name: "input string error",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("input string error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", errors.New("input string error")),
				)
			},
		},
		{
			name: "pull error",
			args: []string{"-v", common.MockAvsPkg.Version(), common.MockAvsPkg.Repo()},
			err:  errors.New("pull error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				d.EXPECT().
					Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{Version: common.MockAvsPkg.Version()}, true).
					Return(daemon.PullResult{}, errors.New("pull error"))
			},
		},
		{
			name: "select profile error",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("select profile error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("", errors.New("select profile error")),
				)
			},
		},
		{
			name: "invalid profile",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("profile invalid-profile not found"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("invalid-profile", nil),
				)
			},
		},
		{
			name: "install error",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("install error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", errors.New("install error")),
				)
			},
		},
		{
			name: "hardware requirements not met",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("hardware requirements not met"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {
									MinCPUCores:                 2,
									MinRAM:                      2048,
									MinFreeSpace:                5120,
									StopIfRequirementsAreNotMet: false,
								},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{
						MinCPUCores:                 2,
						MinRAM:                      2048,
						MinFreeSpace:                5120,
						StopIfRequirementsAreNotMet: false,
					}).Return(false, errors.New("hardware requirements not met")),
				)
			},
		},
		{
			name: "hardware not met and stop",
			args: []string{common.MockAvsPkg.Repo()},
			err:  errors.New("profile profile1 does not meet the hardware requirements"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Options: map[string][]daemon.Option{
								"profile1": {},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {
									MinCPUCores:                 2,
									MinRAM:                      2048,
									MinFreeSpace:                5120,
									StopIfRequirementsAreNotMet: true,
								},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{
						MinCPUCores:                 2,
						MinRAM:                      2048,
						MinFreeSpace:                5120,
						StopIfRequirementsAreNotMet: true,
					}).Return(false, nil),
				)
			},
		},
		{
			name: "commit specified",
			args: []string{"--commit", common.MockAvsPkg.CommitHash(), common.MockAvsPkg.Repo()},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull(common.MockAvsPkg.Repo(), daemon.PullTarget{Commit: common.MockAvsPkg.CommitHash()}, true).
						Return(daemon.PullResult{
							Version: common.MockAvsPkg.Version(),
							Commit:  common.MockAvsPkg.CommitHash(),
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
							HardwareRequirements: map[string]daemon.HardwareRequirements{
								"profile1": {
									MinCPUCores:                 2,
									MinRAM:                      2048,
									MinFreeSpace:                5120,
									StopIfRequirementsAreNotMet: true,
								},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					d.EXPECT().CheckHardwareRequirements(daemon.HardwareRequirements{
						MinCPUCores:                 2,
						MinRAM:                      2048,
						MinFreeSpace:                5120,
						StopIfRequirementsAreNotMet: true,
					}).Return(true, nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().InitMonitoring(false, false).Return(nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     common.MockAvsPkg.Repo(),
							Version: common.MockAvsPkg.Version(),
							Commit:  common.MockAvsPkg.CommitHash(),
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-pkg-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run("mock-avs-pkg-default").Return(nil),
				)
			},
		},
	}

	for _, tc := range ts {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			d := daemonMock.NewMockDaemon(controller)
			p := prompterMock.NewMockPrompter(controller)
			if tc.daemonMock != nil {
				tc.daemonMock(d, p)
			}

			installCmd := InstallCmd(d, p)

			installCmd.SetArgs(tc.args)
			err := installCmd.Execute()

			if tc.err == nil {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err.Error())
			}
		})
	}
}
