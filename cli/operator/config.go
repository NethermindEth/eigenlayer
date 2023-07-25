package operator

import (
	"github.com/NethermindEth/eigenlayer/cli/operator/config"
	"github.com/spf13/cobra"
)

func ConfigCmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "config",
	}

	cmd.AddCommand(
		config.CreateCmd(),
	)

	return &cmd
}
