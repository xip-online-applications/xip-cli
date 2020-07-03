package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *AwsCommands) Identity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Retrieve current identity",
		Run:   c.IdentityRun,
	}

	return cmd
}

func (c *AwsCommands) IdentityRun(cmd *cobra.Command, args []string) {
	fmt.Println(c.Functions.Identity())
}
