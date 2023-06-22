package main

import (
	"log"

	"github.com/NethermindEth/egn/cli"
	"github.com/NethermindEth/egn/cli/prompter"
	"github.com/NethermindEth/egn/internal/commands"
	"github.com/NethermindEth/egn/internal/compose"
	"github.com/NethermindEth/egn/pkg/daemon"
)

func main() {
	// Init docker compose manager
	cmdRunner := commands.NewCMDRunner()
	dockerCompose := compose.NewComposeManager(&cmdRunner)
	// Initialize daemon
	daemon, err := daemon.NewWizDaemon(dockerCompose)
	if err != nil {
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
