package main

import (
	"log"

	"github.com/NethermindEth/eigen-wiz/pkg/cli"
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/NethermindEth/eigen-wiz/pkg/install"
)

func main() {
	// Initialize daemon
	daemon := daemon.NewWizDaemon(install.NewInstaller())
	// Build CLI
	cmd := cli.RootCmd(daemon)
	// Execute CLI
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
