package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Sync() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sync [profile name]",
		Short: "Sync credentials to the credentials file",
		Args:  cobra.RangeArgs(0, 1),
		Run:   SyncRun,
	}

	return cmd
}

func SyncRun(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("config")

	if len(args) == 1 {
		functions.Sync(path, args[0])
	} else {
		for _, value := range functions.GetAllProfileNames(path) {
			functions.Sync(path, value)
		}
	}
}
