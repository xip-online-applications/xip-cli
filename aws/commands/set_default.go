package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Default() *cobra.Command {
	return &cobra.Command{
		Use:   "default [profile]",
		Short: "Set the default profile for use with",
		Args:  cobra.ExactArgs(1),
		Run:   DefaultRun,
	}
}

func DefaultRun(cmd *cobra.Command, args []string) {
	profile := args[0]
	path, _ := cmd.Flags().GetString("config")

	functions.SetDefault(path, profile)
}
