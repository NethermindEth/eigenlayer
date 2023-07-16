package main

import (
	"log"

	"github.com/NethermindEth/eigenlayer/cli"
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/internal/commands"
	"github.com/NethermindEth/eigenlayer/internal/compose"
	"github.com/NethermindEth/eigenlayer/internal/data"
	"github.com/NethermindEth/eigenlayer/internal/docker"
	"github.com/NethermindEth/eigenlayer/internal/locker"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/grafana"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/node_exporter"
	"github.com/NethermindEth/eigenlayer/pkg/monitoring/services/prometheus"
	"github.com/docker/docker/client"
	"github.com/spf13/afero"
)

func main() {
	// Docker client
	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
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

	// Set DataDir
	dataDir, err := data.NewDataDirDefault(fs, locker)
	if err != nil {
		log.Fatal(err)
	}

	// Initialize daemon
	daemon, err := daemon.NewWizDaemon(dataDir, composeManager, dockerManager, monitoringManager, locker)
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
