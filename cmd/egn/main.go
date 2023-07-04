package main

import (
	"log"

	"github.com/NethermindEth/egn/cli"
	"github.com/NethermindEth/egn/cli/prompter"
	"github.com/NethermindEth/egn/internal/commands"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/internal/docker"
	"github.com/NethermindEth/egn/internal/locker"
	"github.com/NethermindEth/egn/internal/monitoring"
	"github.com/NethermindEth/egn/internal/monitoring/services/grafana"
	"github.com/NethermindEth/egn/internal/monitoring/services/node_exporter"
	"github.com/NethermindEth/egn/internal/monitoring/services/prometheus"
	"github.com/NethermindEth/egn/pkg/daemon"
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
)

func main() {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}
	defer dockerClient.Close()

	// Init Cmd runner
	// runner := commands.NewCMDRunnerWithSudo() // Uncomment this line if sudo is required
	runner := commands.NewCMDRunner()

	// Init docker and compose managers
	dockerManager := docker.NewDockerManager(dockerClient)
	composeManager := compose.NewComposeManager(&runner)

	// Set filesystem
	// fs := afero.NewMemMapFs() // Uncomment this line if you want to use the in-memory filesystem
	// fs := afero.NewBasePathFs(afero.NewOsFs(), "/tmp") // Uncomment this line if you want to use the real filesystem with a base path
	fs := afero.NewOsFs() // Uncomment this line if you want to use the real filesystem

	// Set locker
	locker := locker.NewFLock()

	// Get the monitoring manager
	monitoringServices := []monitoring.ServiceAPI{
		grafana.NewGrafana(),
		prometheus.NewPrometheus(),
		node_exporter.NewNodeExporter(),
	}
	monitoringManager := monitoring.NewMonitoringManager(
		monitoringServices,
		composeManager,
		dockerManager,
		fs,
		locker,
	)

	// Initialize daemon
	daemon, err := daemon.NewWizDaemon(composeManager, dockerManager, monitoringManager, fs, locker)
	if err != nil {
		log.Fatal(err)
	}
	if err := daemon.Init(); err != nil {
		log.Fatal(err)
	}

	// Initialize prompter
	p := prompter.NewPrompter()
	// Build CLI
	cmd := cli.RootCmd(daemon, p)
	// Execute CLI
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
