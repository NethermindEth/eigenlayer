package cli

import (
	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func RootCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	cmd := cobra.Command{
		Use:           "eigenlayer",
		SilenceUsage:  true, // Don't show usage when an error occurs
		SilenceErrors: true, // Don't show errors when an error occurs. We handle errors ourselves
	}
	cmd.AddCommand(
		InstallCmd(d, p),
		LocalInstallCmd(d),
		StopCmd(d),
		UninstallCmd(d),
		PluginCmd(d),
		RunCmd(d),
		ListCmd(d),
		LogsCmd(d),
		InitMonitoringCmd(d),
		CleanMonitoringCmd(d),
		UpdateCmd(d, p),
		LocalUpdateCmd(d, p),
		BackupCmd(d),
		RestoreCmd(d),
		OperatorCmd(p),
	)
	cmd.CompletionOptions.DisableDefaultCmd = true
	return &cmd
}
