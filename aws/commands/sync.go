package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Sync() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [profile name]",
		Short: "Sync credentials to the credentials file",
		Args:  cobra.MaximumNArgs(1),
		Run:   SyncRun,
	}

	return cmd
}

func SyncRun(cmd *cobra.Command, args []string) {
	var profile string
	if len(args) > 0 {
		profile = args[0]
	} else {
		profile = functions.GetDefault()
	}

	path, _ := cmd.Flags().GetString("config")

	functions.Sync(path, profile)
}
