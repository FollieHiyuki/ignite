package cmd

import (
	"io"
	"os"

	"github.com/lithammer/dedent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/weaveworks/ignite/cmd/ignite/cmd/cmdutil"
	"github.com/weaveworks/ignite/pkg/config"
	"github.com/weaveworks/ignite/pkg/logs"
	logflag "github.com/weaveworks/ignite/pkg/logs/flag"
	networkflag "github.com/weaveworks/ignite/pkg/network/flag"
	"github.com/weaveworks/ignite/pkg/providers"
	"github.com/weaveworks/ignite/pkg/providers/ignited"
	"github.com/weaveworks/ignite/pkg/util"
	runtimeflag "github.com/weaveworks/ignite/pkg/runtime/flag"
	versioncmd "github.com/weaveworks/ignite/pkg/version/cmd"
)

var logLevel = log.InfoLevel

// Ignite config file path flag variable.
var configPath string

// NewIgnitedCommand returns the root command for ignited
func NewIgnitedCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	root := &cobra.Command{
		Use:   "ignited",
		Short: "ignited: run Firecracker VMs declaratively through a manifest directory or Git",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Set the desired logging level, now that the flags are parsed
			logs.Logger.SetLevel(logLevel)

			if isNonRootCommand(cmd.Name(), cmd.Parent().Name()) {
				return
			}

			// Ignited needs to run as root for now, see
			// https://github.com/weaveworks/ignite/issues/46
			// TODO: Remove this when ready
			util.GenericCheckErr(util.TestRoot())

			// Create the directories needed for running
			util.GenericCheckErr(util.CreateDirectories())

			// Preload necessary providers
			util.GenericCheckErr(providers.Populate(ignited.Preload))

			if err := config.ApplyConfiguration(configPath); err != nil {
				log.Fatal(err)
			}

			// Populate the providers after flags have been parsed
			if err := providers.Populate(ignited.Providers); err != nil {
				log.Fatal(err)
			}
		},
		Long: dedent.Dedent(`
			Ignite is a containerized Firecracker microVM administration tool.
			It can build VM images, spin VMs up/down and manage multiple VMs efficiently.

			TODO: ignited documentation
		`),
	}

	addGlobalFlags(root.PersistentFlags())

	root.AddCommand(NewCmdCompletion(os.Stdout, root))
	root.AddCommand(NewCmdGitOps(os.Stdout))
	root.AddCommand(NewCmdDaemon(os.Stdout))
	root.AddCommand(versioncmd.NewCmdVersion(os.Stdout))
	return root
}

func isNonRootCommand(cmd string, parentCmd string) bool {
	if parentCmd != "ignited" {
		return false
	}

	switch cmd {
	case "version", "help", "completion":
		return true
	}

	return false
}

func addGlobalFlags(fs *pflag.FlagSet) {
	logflag.LogLevelFlagVar(fs, &logLevel)
	runtimeflag.RuntimeVar(fs, &providers.RuntimeName)
	networkflag.NetworkPluginVar(fs, &providers.NetworkPluginName)
	cmdutil.AddIDPrefixFlag(fs, &providers.IDPrefix)
	fs.StringVar(&configPath, "ignite-config", "", "Ignite configuration path; refer to the 'Ignite Configuration' docs for more details")
}
