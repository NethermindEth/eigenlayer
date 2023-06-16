package cli

import (
	"github.com/NethermindEth/eigen-wiz/internal/prompter"
	"github.com/NethermindEth/eigen-wiz/pkg/daemon"
	"github.com/spf13/cobra"
)

func InstallCmd(d daemon.Daemon) *cobra.Command {
	var (
		url     string
		version string
		tag     string
	)
	cmd := cobra.Command{
		Use:   "install [URL]",
		Short: "Install AVS node software from a git repository",
		Long:  "Installs the AVS node software, downloading it from a git repository. You will need to provide the repository URL as a unique argument, which must be an HTTP or HTTPS URL. Use the --version flag if you need to specify a version.",
		Args:  cobra.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			url = args[0]
			return validatePkgURL(url)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return d.Install(daemon.InstallOptions{
				URL:     url,
				Version: version,
				Tag:     tag,
				ProfileSelector: func(profiles []string) (string, error) {
					return prompter.SelectProfile(profiles)
				},
				OptionsFiller: func(opts []daemon.Option) (outOpts []daemon.Option, err error) {
					outOpts = make([]daemon.Option, len(opts))
					for i, o := range opts {
						_, err = prompter.InputString(o.Name(), o.Default(), o.Help(), func(s string) error {
							return o.Set(s)
						})
						if err != nil {
							break
						}
						outOpts[i] = o
					}
					return
				},
				RunConfirmation: func() (bool, error) {
					return prompter.Confirm("Run the node now?")
				},
			})
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to install. If not specified, the latest version will be installed.")
	cmd.Flags().StringVarP(&tag, "tag", "t", "default", "tag to use for the new instance name")
	return &cmd
}
