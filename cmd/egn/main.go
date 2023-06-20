package main

import (
	"log"

	"github.com/NethermindEth/egn/cli"
	"github.com/NethermindEth/egn/pkg/daemon"
)

func main() {
	// Initialize daemon
	daemon, err := daemon.NewWizDaemon()
	if err != nil {
		log.Fatal(err)
	}
	// Build CLI
	cmd := cli.RootCmd(daemon)
	// Execute CLI
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
