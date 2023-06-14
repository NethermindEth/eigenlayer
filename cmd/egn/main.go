package main

import (
	"log"

	"github.com/NethermindEth/eigen-wiz/cli"
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
)

func main() {
	// Initialize daemon
	daemon := daemon.NewWizDaemon()
	// Build CLI
	cmd := cli.RootCmd(daemon)
	// Execute CLI
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
