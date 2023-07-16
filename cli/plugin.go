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
	)
	cmd := cobra.Command{
		Use:  "plugin [FLAGS] [INSTANCE_ID] [PLUGIN ARGS]",
		Long: `Run a plugin. The instance id is required as the unique argument. The plugin arguments are passed to the plugin as is.`,
		Args: cobra.MinimumNArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			instanceId = args[0]
			if !d.HasInstance(instanceId) {
				return errors.New("instance not found")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.RunPlugin(instanceId, args[1:], noDestroyImage)
		},
	}

	cmd.Flags().BoolVar(&noDestroyImage, "no-rm-image", false, "Do not remove the plugin image after plugin execution")
	cmd.DisableFlagParsing = true // Flag parsing is disable to support dynamic flags for plugin arguments

	return &cmd
}
