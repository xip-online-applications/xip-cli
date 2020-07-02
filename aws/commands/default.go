package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

func (c *AwsCommands) Default() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default [profile]",
		Short: "Retrieve the current default profile or set the default by providing the profile.",
		Args:  cobra.RangeArgs(0, 1),
		Run:   c.DefaultRun,
	}

	return cmd
}

func (c *AwsCommands) DefaultRun(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		c.Functions.SetDefault(args[0])
	}

	prof, err := c.Functions.GetDefaultProfile()
	if err != nil {
		panic(err)
	}

	fmt.Println(prof)
}
