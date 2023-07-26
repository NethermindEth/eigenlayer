package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	"github.com/spf13/cobra"
)

var volumeNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]+$`)

func PluginCmd(d daemon.Daemon) *cobra.Command {
	var (
		instanceId     string
		noDestroyImage bool
		host           bool
		pluginArgs     []string
		volumes        []string
	)
	cmd := cobra.Command{
		Use:     "plugin [flags] [instance_id] [plugin_args]",
		Short:   "Run an AVS node plugin",
		Long:    `Run a plugin. The instance id is required as the unique argument. The plugin arguments are passed to the plugin as is.`,
		Args:    cobra.MinimumNArgs(1),
		Example: `
- Basic usage:

	$ eigenlayer plugin mock-avs-default

  In this case the plugin will run on the AVS network and will receive no
  no arguments and no volumes.

- Using the host network:

	$ eigenlayer plugin --host mock-avs-default --host localhost --port 8081

  In this case the plugin will run on the host network and will receive the
  following arguments: '--hot localhost --port 8081'.

- Using volumes:
	
	$ eigenlayer plugin --volume /tmp:/tmp --volume plugin-v:/data mock-avs-default

  This will mount the /tmp directory of the host inside the plugin container at 
  /tmp, and the plugin-v volume at /data.
`,

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
			runPluginOptions := daemon.RunPluginOptions{
				NoDestroyImage: noDestroyImage,
				HostNetwork:    host,
				Volumes:        make(map[string]string),
				Binds:          make(map[string]string),
			}
			for _, v := range volumes {
				vSplit := strings.Split(v, ":")
				if len(vSplit) != 2 {
					return fmt.Errorf("invalid volume format: %s", v)
				}
				if volumeNameRegex.MatchString(vSplit[0]) {
					// Add mount of type volume
					runPluginOptions.Volumes[vSplit[0]] = vSplit[1]
				} else {
					fStat, err := os.Stat(filepath.Clean(vSplit[0]))
					if err != nil {
						if errors.Is(err, os.ErrNotExist) {
							return fmt.Errorf("dir does not exist: %s", vSplit[0])
						}
						return fmt.Errorf("failed to stat volume %s: %s", args[0], err)
					}
					if !fStat.IsDir() {
						return fmt.Errorf("volume is not a directory: %s", vSplit[0])
					}
					// Add mount of type bind
					runPluginOptions.Binds[vSplit[0]] = vSplit[1]
				}
			}
			return d.RunPlugin(instanceId, pluginArgs, runPluginOptions)
		},
	}

	cmd.Flags().BoolVar(&noDestroyImage, "no-rm-image", false, "Do not remove the plugin image after plugin execution")
	cmd.Flags().BoolVar(&host, "host", false, "Run the plugin on the host network instead of the AVS network")
	cmd.Flags().StringSliceVarP(&volumes, "volume", "v", []string{}, "Bind mount a volume. Format: <volume_name>:<path> or <path>:<path>. Can be specified multiple times")
	cmd.Flags().SetInterspersed(false)
	return &cmd
}
