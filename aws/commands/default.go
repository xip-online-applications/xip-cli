package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Default() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default [profile]",
		Short: "Retrieve the current default profile or set the default by providing the profile.",
		Args:  cobra.RangeArgs(0, 1),
		Run:   DefaultRun,
	}

	return cmd
}

func DefaultRun(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		profile := args[0]
		path, _ := cmd.Flags().GetString("config")

		functions.SetDefault(path, profile)
	}

	fmt.Println(functions.GetDefaultProfile())
}
