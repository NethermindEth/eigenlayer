package cli

import (
	"errors"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

func PluginCmd(d daemon.Daemon) *cobra.Command {
	var (
		instanceId     string
		noDestroyImage bool
		host           bool
		pluginArgs     []string
	)
	cmd := cobra.Command{
		Use:   "plugin [flags] [instance_id] [plugin_args]",
		Short: "Run an AVS node plugin",
		Long:  `Run a plugin. The instance id is required as the unique argument. The plugin arguments are passed to the plugin as is.`,
		Args:  cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			instanceId = args[0]
			if !d.HasInstance(instanceId) {
				return errors.New("instance not found")
			}
			if len(args) > 1 {
				pluginArgs = args[1:]
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunPlugin(instanceId, pluginArgs, daemon.RunPluginOptions{
				NoDestroyImage: noDestroyImage,
				HostNetwork:    host,
			})
		},
	}

	cmd.Flags().BoolVar(&noDestroyImage, "no-rm-image", false, "Do not remove the plugin image after plugin execution")
	cmd.Flags().BoolVar(&host, "host", false, "Run the plugin on the host network instead of the AVS network")
	cmd.Flags().SetInterspersed(false)
	return &cmd
}
