package compose

import (
	"fmt"
	"strings"

	"github.com/NethermindEth/eigen-wiz/internal/commands"
)

// DockerComposeCmdError represents an error that occurs when running a Docker Compose command.
type DockerComposeCmdError struct {
	cmd string
}

// Error returns a string representation of the DockerComposeCmdError.
func (e DockerComposeCmdError) Error() string {
	return fmt.Sprintf("docker compose manager running docker compose %s", e.cmd)
}

// CMDRunner is an interface that defines a method for running commands.
type CMDRunner interface {
	RunCMD(commands.Command) (string, int, error)
}

// ComposeManager manages Docker Compose operations.
type ComposeManager struct {
	cmdRunner CMDRunner
}

// NewComposeManager creates a new instance of ComposeManager.
func NewComposeManager(runner CMDRunner) ComposeManager {
	return ComposeManager{
		cmdRunner: runner,
	}
}

// Up runs the Docker Compose 'up' command for the specified options.
func (cm *ComposeManager) Up(opts DockerComposeUpOptions) error {
	upCmd := fmt.Sprintf("docker compose -f %s up -d", opts.Path)
	if len(opts.Services) > 0 {
		upCmd += " " + strings.Join(opts.Services, " ")
	}

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: upCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "up"}, err)
	}
	return nil
}

// Pull runs the Docker Compose 'pull' command for the specified options.
func (cm *ComposeManager) Pull(opts DockerComposePullOptions) error {
	pullCmd := fmt.Sprintf("docker compose -f %s pull", opts.Path)
	if len(opts.Services) > 0 {
		pullCmd += " " + strings.Join(opts.Services, " ")
	}

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: pullCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "pull"}, err)
	}
	return nil
}

// Create runs the Docker Compose 'create' command for the specified options.
func (cm *ComposeManager) Create(opts DockerComposeCreateOptions) error {
	createCmd := fmt.Sprintf("docker compose -f %s create", opts.Path)
	if len(opts.Services) > 0 {
		createCmd += " " + strings.Join(opts.Services, " ")
	}

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: createCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "create"}, err)
	}
	return nil
}

// Build runs the Docker Compose 'build' command for the specified options.
func (cm *ComposeManager) Build(opts DockerComposeBuildOptions) error {
	buildCmd := fmt.Sprintf("docker compose -f %s build", opts.Path)
	if len(opts.Services) > 0 {
		buildCmd += " " + strings.Join(opts.Services, " ")
	}

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: buildCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "build"}, err)
	}
	return nil
}

// PS runs the Docker Compose 'ps' command for the specified options and returns the output.
func (cm *ComposeManager) PS(opts DockerComposePsOptions) (string, error) {
	var psCmd string
	if opts.Path != "" {
		psCmd = fmt.Sprintf("docker compose -f %s ps", opts.Path)
	} else {
		psCmd = "docker compose ps"
	}
	if opts.Services {
		psCmd += " --services"
	}
	if opts.Quiet {
		psCmd += " --quiet"
	}
	if opts.FilterRunning {
		psCmd += " --filter status=running"
	}
	if opts.ServiceName != "" {
		psCmd += " " + opts.ServiceName
	}

	out, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: psCmd, GetOutput: true})
	if err != nil {
		return "", fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "ps"}, err)
	}
	return out, nil
}

// Logs runs the Docker Compose 'logs' command for the specified options.
func (cm *ComposeManager) Logs(opts DockerComposeLogsOptions) error {
	logsCmd := fmt.Sprintf("docker compose -f %s logs", opts.Path)
	if opts.Follow {
		logsCmd += fmt.Sprintf(" --follow")
	}
	if opts.Tail > 0 {
		logsCmd += fmt.Sprintf(" --tail=%d", opts.Tail)
	}
	if len(opts.Services) > 0 {
		logsCmd += " " + strings.Join(opts.Services, " ")
	}

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: logsCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "logs"}, err)
	}
	return nil
}

// Down runs the Docker Compose 'down' command for the specified options.
func (cm *ComposeManager) Down(opts DockerComposeDownOptions) error {
	downCmd := fmt.Sprintf("docker compose -f %s down", opts.Path)

	if _, _, err := cm.cmdRunner.RunCMD(commands.Command{Cmd: downCmd}); err != nil {
		return fmt.Errorf("%w: %s", DockerComposeCmdError{cmd: "down"}, err)
	}
	return nil
}
