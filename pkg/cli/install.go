package cli

import (
	"fmt"
	"os"

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
			destDir, err := os.MkdirTemp(os.TempDir(), "egn-install")
			if err != nil {
				return err
			}
			fmt.Println("Pulling package...")
			pullResponse, err := d.Pull(&daemon.PullOptions{
				URL:     url,
				Version: version,
				DestDir: destDir,
			})
			if err != nil {
				return err
			}
			fmt.Println("Package pulled successfully.")
			profileNames := make([]string, 0, len(pullResponse.Profiles))
			for k := range pullResponse.Profiles {
				profileNames = append(profileNames, k)
			}
			selectedProfile, err := prompter.SelectProfile(profileNames)
			if err != nil {
				return err
			}
			fmt.Printf("Selected profile: %s\n", selectedProfile)
			profileOptions := pullResponse.Profiles[selectedProfile]
			for _, option := range profileOptions {
				_, err := prompter.InputString(option.Name(), option.Default(), option.Help(), func(s string) error {
					return option.Set(s)
				})
				if err != nil {
					return err
				}
			}
			fmt.Printf("%+v\n", profileOptions)
			return nil
		},
	}

	cmd.Flags().StringVarP(&version, "version", "v", "", "version to install. If not specified, the latest version will be installed.")
	cmd.Flags().StringVarP(&tag, "tag", "t", "default", "tag to use for the new instance name")
	return &cmd
}
