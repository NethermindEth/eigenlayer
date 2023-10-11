package cli

import (
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"text/tabwriter"

	"github.com/NethermindEth/eigenlayer/cli/prompter"
	"github.com/NethermindEth/eigenlayer/pkg/daemon"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/mod/semver"
)

var hashRegex = regexp.MustCompile("^[0-9a-f]{40}$")

func UpdateCmd(d daemon.Daemon, p prompter.Prompter) *cobra.Command {
	var (
		instanceId string
		version    string
		commit     string
		noPrompt   bool
		backup     bool
		help       bool
		yes        bool
	)
	cmd := cobra.Command{
		Use:   "update [flags] <instance_id> <version>",
		Short: "Update an instance to a new version.",
		Long: `Updates instance <instance_id> to a new version using the specified
version or commit hash in the <version> argument. If no version is specified, the
latest version will be used. If the new version is lower or equal to the current
version, the update will fail. Also, if the new commit passed as argument or the
commit of the new version is not a descendant of the current commit, the update
will fail.

The update process will check if the new version has the same profile as the
current one. If it's not the case, the process will fail. If is necessary to change
profiles between versions, the user should install the new version and profile
using the install command and then uninstall the current version.

Options of the new version can be specified using the --option.<option-name> flag.

To avoid any data loss during the update process, the user can specify the --backup
flag. In this case, the current instance will be backed up before uninstalling it,
and if the update process fails, the instance will be restored. Also, the backup
could be restored manually using the 'eigenlayer restore' command.`,
		Example: `
- Updating to the latest version:
	
	$ eigenlayer update mock-avs-default

  In this case the latest version of the package will be pulled and tried to be
  installed.

- Updating to a specific version:

	$ eigenlayer update mock-avs-default v5.5.0

  In this case the version v5.5.0 of the package will be pulled and tried to be
  installed.

- Updating to a specific commit:

    $ eigenlayer update mock-avs-default 3b2c50c15e53ae7afebbdbe210b834d1ee471043

  In this case the commit 3b2c50c15e53ae7afebbdbe210b834d1ee471043 of the package
  will be pulled and tried to be installed.
`,
		DisableFlagParsing: true,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			// Parse static flags
			cmd.DisableFlagParsing = false
			cmd.FParseErrWhitelist.UnknownFlags = true // Don't show error for unknown flags to allow dynamic flags
			err := cmd.ParseFlags(args)
			if err != nil {
				return err
			}

			// Skip execution if help flag is set
			help, err = cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if help {
				return nil
			}

			// Validate args
			args = cmd.Flags().Args()
			if len(args) < 1 {
				return fmt.Errorf("%w: instance-id is required", ErrInvalidNumberOfArgs)
			}
			if len(args) > 2 {
				return fmt.Errorf("%w: too many arguments", ErrInvalidNumberOfArgs)
			}
			if len(args) >= 1 {
				instanceId = args[0]
			}
			if len(args) == 2 {
				if semver.IsValid(args[1]) {
					version = args[1]
				} else if hashRegex.MatchString(args[1]) {
					commit = args[1]
				} else {
					return fmt.Errorf("%w: invalid version or commit", ErrInvalidArgs)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if help {
				return cmd.Help()
			}
			// Pull update
			pullResult, err := pullUpdate(d, instanceId, version, commit)
			if err != nil {
				if errors.Is(err, daemon.ErrVersionAlreadyInstalled) {
					log.Info(err.Error())
					return nil
				}
				return err
			}

			// Log useful information about the update
			logVersionChange(pullResult.OldVersion, pullResult.NewVersion)
			logCommitChange(pullResult.OldCommit, pullResult.NewCommit)
			printOptionsTable(pullResult.OldOptions, pullResult.MergedOptions)

			// Build dynamic flags with the profile options
			for _, o := range pullResult.MergedOptions {
				v, err := o.Value()
				if err != nil {
					cmd.Flags().String("option."+o.Name(), o.Default(), o.Help())
				} else {
					cmd.Flags().String("option."+o.Name(), v, o.Help())
				}
			}

			// Parse dynamic flags
			cmd.FParseErrWhitelist.UnknownFlags = false
			if err = cmd.ParseFlags(args); err != nil {
				return err
			}

			// Fill necessary options
			for _, o := range pullResult.MergedOptions {
				if noPrompt {
					flagValue, err := cmd.Flags().GetString("option." + o.Name())
					if err != nil {
						return err
					}
					if flagValue == "" {
						return fmt.Errorf("%w: %s", ErrOptionWithoutDefault, o.Name())
					}
					if err = o.Set(flagValue); err != nil {
						return err
					}
				}
				if !o.IsSet() {
					_, err := p.InputString(o.Name(), o.Default(), o.Help(), func(s string) error {
						return o.Set(s)
					})
					if err != nil {
						return err
					}
				}
			}

			// Backup instance
			var backupId string
			if backup {
				backupId, err = d.Backup(instanceId)
				if err != nil {
					return err
				}
				log.Info("Backup created with id: ", backupId)
			}

			// Uninstall current instance
			err = uninstallPackage(d, instanceId)
			if err != nil {
				if backup {
					return abortWithRestore(d, backupId, err)
				}
				return err
			}

			// Install new instance's version
			newInstanceId, err := install(d, daemon.InstallOptions{
				Name:    pullResult.Name,
				Tag:     pullResult.Tag,
				URL:     pullResult.Url,
				Version: pullResult.NewVersion,
				Commit:  pullResult.NewCommit,
				Profile: pullResult.Profile,
				Options: pullResult.MergedOptions,
			})
			if err != nil {
				if backup {
					return abortWithRestore(d, backupId, err)
				}
				return err
			}
			if newInstanceId != instanceId {
				// NOTE: I think this never happens but it could be useful to check
				// that the instance ID is the same as the one we started with. Also
				// we can manage this case as an error, but I think it's better to
				// just log it for now.
				log.Infof("Instance ID changed: %s -> %s", instanceId, newInstanceId)
			}

			if pullResult.HasPlugin {
				log.Info("The installed node software has a plugin.")
			}

			return runInstance(d, newInstanceId, p, yes, noPrompt)
		},
	}

	cmd.Flags().BoolVar(&noPrompt, "no-prompt", false, "disable command prompts, and all options should be passed using command flags.")
	cmd.Flags().BoolVarP(&yes, "yes", "y", false, "skip confirmation prompts.")
	cmd.Flags().BoolVar(&backup, "backup", false, "backup current instance before updating.")
	return &cmd
}

func abortWithRestore(d daemon.Daemon, backupId string, updateErr error) error {
	log.Errorf("Update process failed with error: %s", updateErr.Error())
	log.Infof("Restoring instance from backup %s...", backupId)
	return d.Restore(backupId, false)
}

func pullUpdate(d daemon.Daemon, instanceID, version, commit string) (daemon.PullUpdateResult, error) {
	log.Info("Pulling package...")
	pullResult, err := d.PullUpdate(instanceID, daemon.PullTarget{Version: version, Commit: commit})
	if err == nil {
		log.Info("Package pulled successfully")
	}
	return pullResult, err
}

func uninstallPackage(d daemon.Daemon, instanceID string) error {
	log.Info("Uninstalling current package...")
	err := d.Uninstall(instanceID)
	if err == nil {
		log.Info("Package uninstalled successfully")
	}
	return err
}

func install(d daemon.Daemon, options daemon.InstallOptions) (string, error) {
	log.Info("Installing new package...")
	newInstanceId, err := d.Install(options)
	if err == nil {
		log.Infof("Package installed successfully with instance ID: %s", newInstanceId)
	}
	return newInstanceId, err
}

func runInstance(d daemon.Daemon, instanceID string, p prompter.Prompter, yes, noPrompt bool) error {
	var err error
	if !yes && !noPrompt {
		yes, err = p.Confirm("Run the new instance now?")
		if err != nil {
			return err
		}
	}
	if yes {
		log.Infof("Running instance %s ...", instanceID)
		err = d.Run(instanceID)
	}
	if err == nil {
		log.Infof("Instance %s running successfully", instanceID)
	}
	return err
}

func logVersionChange(oldVersion, newVersion string) {
	if newVersion != "" {
		log.Infof("Package version changed: %s -> %s", oldVersion, newVersion)
	}
}

func logCommitChange(oldCommit, newCommit string) {
	if newCommit != "" {
		log.Infof("Package commit changed from %s -> %s", oldCommit, newCommit)
	}
}

type tableOptionItem struct {
	name string
	old  string
	new  string
}

func (i tableOptionItem) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t", i.name, i.old, i.new)
}

func printOptionsTable(old, merged []daemon.Option) error {
	rows := make(map[string]*tableOptionItem)
	for _, o := range old {
		if o.IsSet() {
			v, err := o.Value()
			if err != nil {
				return err
			}
			if item, ok := rows[o.Name()]; ok {
				item.old = v
			} else {
				rows[o.Name()] = &tableOptionItem{name: o.Name(), old: v, new: "<deprecated>"}
			}
		} else {
			if _, ok := rows[o.Name()]; !ok {
				rows[o.Name()] = &tableOptionItem{name: o.Name(), old: "<not set>"}
			}
		}
	}
	for _, o := range merged {
		if o.IsSet() {
			v, err := o.Value()
			if err != nil {
				return err
			}
			if item, ok := rows[o.Name()]; ok {
				item.new = v
			} else {
				rows[o.Name()] = &tableOptionItem{name: o.Name(), old: "<not set>", new: v}
			}
		} else {
			if v, ok := rows[o.Name()]; ok {
				v.new = "<to be set>"
			} else {
				rows[o.Name()] = &tableOptionItem{name: o.Name(), old: "<not set>", new: "<to be set>"}
			}
		}
	}
	var out bytes.Buffer
	w := tabwriter.NewWriter(&out, 0, 0, 4, ' ', 0)
	fmt.Fprintln(w, "OPTION NAME\tOLD VALUE\tNEW VALUE\t")
	for _, row := range rows {
		fmt.Fprintln(w, row)
	}
	w.Flush()
	log.Debug(out.String())
	return nil
}
