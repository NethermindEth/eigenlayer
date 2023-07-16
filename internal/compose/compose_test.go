package compose

import (
	"errors"
	"strconv"
	"strings"
	"testing"

	"github.com/NethermindEth/eigenlayer/internal/commands"
	"github.com/NethermindEth/eigenlayer/internal/compose/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestUp(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeUpOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposeUpOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeUpOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "up"},
		},
		{
			name: "it runs the correct command when no services are specified",
			opts: DockerComposeUpOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{},
			},
			runCMDError: nil,
			wantError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			var expectedCmd string
			if len(tt.opts.Services) > 0 {
				expectedCmd = "docker compose -f " + tt.opts.Path + " up -d " + strings.Join(tt.opts.Services, " ")
			} else {
				expectedCmd = "docker compose -f " + tt.opts.Path + " up -d"
			}

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Up(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Up() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Up command
	opts := DockerComposeUpOptions{
		Path:     "/path/to/docker-compose.yml",
		Services: []string{"service1", "service2"},
	}

	// Run the Docker Compose Up command
	manager.Up(opts)
}

func TestPull(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposePullOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposePullOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposePullOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "pull"},
		},
		{
			name: "it runs the correct command when no services are specified",
			opts: DockerComposePullOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{},
			},
			runCMDError: nil,
			wantError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			var expectedCmd string
			if len(tt.opts.Services) > 0 {
				expectedCmd = "docker compose -f " + tt.opts.Path + " pull " + strings.Join(tt.opts.Services, " ")
			} else {
				expectedCmd = "docker compose -f " + tt.opts.Path + " pull"
			}

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Pull(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Pull() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Pull command
	opts := DockerComposePullOptions{
		Path:     "/path/to/docker-compose.yml",
		Services: []string{"service1", "service2"},
	}

	// Run the Docker Compose Pull command
	manager.Pull(opts)
}

func TestCreate(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeCreateOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposeCreateOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
				Build:    true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeCreateOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "create"},
		},
		{
			name: "it runs the correct command when no services are specified",
			opts: DockerComposeCreateOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{},
			},
			runCMDError: nil,
			wantError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			expectedCmd := "docker compose -f " + tt.opts.Path + " create"
			if tt.opts.Build {
				expectedCmd += " --build"
			}

			if len(tt.opts.Services) > 0 {
				expectedCmd += " " + strings.Join(tt.opts.Services, " ")
			}

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Create(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Create() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Create command
	opts := DockerComposeCreateOptions{
		Path:     "/path/to/docker-compose.yml",
		Services: []string{"service1", "service2"},
	}

	// Run the Docker Compose Create command
	manager.Create(opts)
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeBuildOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposeBuildOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeBuildOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "build"},
		},
		{
			name: "it runs the correct command when no services are specified",
			opts: DockerComposeBuildOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{},
			},
			runCMDError: nil,
			wantError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			var expectedCmd string
			if len(tt.opts.Services) > 0 {
				expectedCmd = "docker compose -f " + tt.opts.Path + " build " + strings.Join(tt.opts.Services, " ")
			} else {
				expectedCmd = "docker compose -f " + tt.opts.Path + " build"
			}

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Build(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Build() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Build command
	opts := DockerComposeBuildOptions{
		Path:     "/path/to/docker-compose.yml",
		Services: []string{"service1", "service2"},
	}

	// Run the Docker Compose Build command
	manager.Build(opts)
}

func TestPS(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposePsOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command with all options set",
			opts: DockerComposePsOptions{
				Path:          "/path/to/docker-compose.yml",
				Services:      true,
				Quiet:         true,
				FilterRunning: true,
				ServiceName:   "service1",
				Format:        "format",
				All:           true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only Services option set",
			opts: DockerComposePsOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only Quiet option set",
			opts: DockerComposePsOptions{
				Path:  "/path/to/docker-compose.yml",
				Quiet: true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only FilterRunning option set",
			opts: DockerComposePsOptions{
				Path:          "/path/to/docker-compose.yml",
				FilterRunning: true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only ServiceName option set",
			opts: DockerComposePsOptions{
				Path:        "/path/to/docker-compose.yml",
				ServiceName: "service1",
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name:        "it runs the correct command with no options or path set",
			opts:        DockerComposePsOptions{},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposePsOptions{
				Path:          "/path/to/docker-compose.yml",
				Services:      true,
				Quiet:         true,
				FilterRunning: true,
				ServiceName:   "service1",
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "ps"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			var expectedCmd string
			if tt.opts.Path != "" {
				expectedCmd = "docker compose -f " + tt.opts.Path + " ps"
			} else {
				expectedCmd = "docker compose ps"
			}

			if tt.opts.Services {
				expectedCmd += " --services"
			}
			if tt.opts.Quiet {
				expectedCmd += " --quiet"
			}
			if tt.opts.FilterRunning {
				expectedCmd += " --filter status=running"
			}
			if tt.opts.Format != "" {
				expectedCmd += " --format " + tt.opts.Format
			}
			if tt.opts.All {
				expectedCmd += " -a"
			}
			if tt.opts.ServiceName != "" {
				expectedCmd += " " + tt.opts.ServiceName
			}

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			_, err := manager.PS(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_PS() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose PS command
	opts := DockerComposePsOptions{
		Path:          "/path/to/docker-compose.yml",
		Services:      true,
		Quiet:         true,
		FilterRunning: true,
		ServiceName:   "service1",
	}

	// Run the Docker Compose PS command
	manager.PS(opts)
}

func TestLogs(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeLogsOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command with all options set",
			opts: DockerComposeLogsOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
				Follow:   true,
				Tail:     10,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only Follow option set",
			opts: DockerComposeLogsOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
				Follow:   true,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it runs the correct command with only Tail option set",
			opts: DockerComposeLogsOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
				Tail:     10,
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeLogsOptions{
				Path:     "/path/to/docker-compose.yml",
				Services: []string{"service1", "service2"},
				Follow:   true,
				Tail:     10,
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "logs"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			expectedCmd := "docker compose -f " + tt.opts.Path + " logs"
			if tt.opts.Follow {
				expectedCmd += " --follow"
			}
			if tt.opts.Tail > 0 {
				expectedCmd += " --tail=" + strconv.Itoa(tt.opts.Tail)
			}
			expectedCmd += " " + strings.Join(tt.opts.Services, " ")

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Logs(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Logs() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Logs command
	opts := DockerComposeLogsOptions{
		Path:     "/path/to/docker-compose.yml",
		Services: []string{"service1", "service2"},
		Follow:   true,
		Tail:     10,
	}

	// Run the Docker Compose Logs command
	manager.Logs(opts)
}

func TestStop(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeStopOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposeStopOptions{
				Path: "/path/to/docker-compose.yml",
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeStopOptions{
				Path: "/path/to/docker-compose.yml",
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "stop"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			expectedCmd := "docker compose -f " + tt.opts.Path + " stop"

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Stop(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDown(t *testing.T) {
	tests := []struct {
		name        string
		opts        DockerComposeDownOptions
		runCMDError error
		wantError   error
	}{
		{
			name: "it runs the correct command",
			opts: DockerComposeDownOptions{
				Path: "/path/to/docker-compose.yml",
			},
			runCMDError: nil,
			wantError:   nil,
		},
		{
			name: "it returns an error if RunCMD fails",
			opts: DockerComposeDownOptions{
				Path: "/path/to/docker-compose.yml",
			},
			runCMDError: errors.New("command failed"),
			wantError:   DockerComposeCmdError{cmd: "down"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockRunner := mocks.NewMockCMDRunner(ctrl)

			manager := NewComposeManager(mockRunner)

			expectedCmd := "docker compose -f " + tt.opts.Path + " down"

			if tt.runCMDError != nil {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 1, tt.runCMDError)
			} else {
				mockRunner.EXPECT().RunCMD(commands.Command{Cmd: expectedCmd, GetOutput: true}).Return("", 0, nil)
			}

			err := manager.Down(tt.opts)

			if tt.wantError != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.wantError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func ExampleComposeManager_Down() {
	// Create a new CMDRunner with admin privileges
	cmdRunner := commands.NewCMDRunnerWithSudo()

	// Create a new ComposeManager with the CMDRunner
	manager := NewComposeManager(&cmdRunner)

	// Define the options for the Docker Compose Down command
	opts := DockerComposeDownOptions{
		Path: "/path/to/docker-compose.yml",
	}

	// Run the Docker Compose Down command
	manager.Down(opts)
}
