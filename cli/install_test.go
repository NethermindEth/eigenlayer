package cli

import (
	"errors"
	"fmt"
	"testing"

	daemonMock "github.com/NethermindEth/egn/cli/mocks"
	prompterMock "github.com/NethermindEth/egn/cli/prompter/mocks"
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
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
			name: "valid arguments, and run",
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(true, nil),
					d.EXPECT().Run("mock-avs-default").Return(nil),
				)
			},
		},
		{
			name: "valid arguments, not run",
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(false, nil),
				)
			},
		},
		{
			name: "input string error",
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("input string error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", errors.New("input string error")),
				)
			},
		},
		{
			name: "pull error",
			args: []string{"-v", "v2.0.2", "https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("pull error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				d.EXPECT().
					Pull("https://github.com/NethermindEth/mock-avs", "v2.0.2", true).
					Return(daemon.PullResult{}, errors.New("pull error"))
			},
		},
		{
			name: "select profile error",
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("select profile error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
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
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("profile invalid-profile not found"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
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
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("install error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", errors.New("install error")),
				)
			},
		},
		{
			name: "run confirm error",
			args: []string{"https://github.com/NethermindEth/mock-avs"},
			err:  errors.New("confirm error"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
					p.EXPECT().Confirm("Run the new instance now?").Return(false, errors.New("confirm error")),
				)
			},
		},
		{
			name: "--run flag",
			args: []string{"--run", "https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Default().Return("default1").Times(2)
				option.EXPECT().Help().Return("help1").Times(2)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					p.EXPECT().Select("Select a profile", []string{"profile1"}).Return("profile1", nil),
					p.EXPECT().InputString("option1", "default1", "help1", gomock.Any()).Return("value1", nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
					d.EXPECT().Run("mock-avs-default").Return(nil),
				)
			},
		},
		{
			name: "--no-prompt and --run",
			args: []string{"--no-prompt", "--run", "--profile", "profile1", "https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(2)
				option.EXPECT().Default().Return("default1")
				option.EXPECT().Help().Return("help1")
				option.EXPECT().Set("default1").Return(nil)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
					d.EXPECT().Run("mock-avs-default").Return(nil),
				)
			},
		},
		{
			name: "--no-prompt and --run false by default",
			args: []string{"--no-prompt", "--profile", "profile1", "https://github.com/NethermindEth/mock-avs"},
			err:  nil,
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(2)
				option.EXPECT().Default().Return("default1")
				option.EXPECT().Help().Return("help1")
				option.EXPECT().Set("default1").Return(nil)

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
					d.EXPECT().
						Install(daemon.InstallOptions{
							URL:     "https://github.com/NethermindEth/mock-avs",
							Version: "v2.0.2",
							Profile: "profile1",
							Options: []daemon.Option{option},
							Tag:     "default",
						}).Return("mock-avs-default", nil),
				)
			},
		},
		{
			name: "--no-prompt with an option without default value",
			args: []string{"--profile=profile1", "--no-prompt", "--run=false", "https://github.com/NethermindEth/mock-avs"},
			err:  fmt.Errorf("%w: %s", ErrOptionWithoutDefault, "option1"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(3)
				option.EXPECT().Help().Return("help1")
				option.EXPECT().Default().Return("")

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
				)
			},
		},
		{
			name: "--no-prompt and set option error",
			args: []string{"--profile=profile1", "--no-prompt", "https://github.com/NethermindEth/mock-avs", "--option.option1", "value1"},
			err:  errors.New("error setting option"),
			daemonMock: func(d *daemonMock.MockDaemon, p *prompterMock.MockPrompter) {
				option := daemonMock.NewMockOption(gomock.NewController(t))
				option.EXPECT().Name().Return("option1").Times(2)
				option.EXPECT().Help().Return("help1")
				option.EXPECT().Default().Return("")
				option.EXPECT().Set("value1").Return(errors.New("error setting option"))

				gomock.InOrder(
					d.EXPECT().
						Pull("https://github.com/NethermindEth/mock-avs", "", true).
						Return(daemon.PullResult{
							Version: "v2.0.2",
							Options: map[string][]daemon.Option{
								"profile1": {option},
							},
						}, nil),
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
