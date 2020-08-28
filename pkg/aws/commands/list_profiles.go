package commands

import (
	"github.com/spf13/cobra"
)

func (c *AwsCommands) ListProfiles() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the registered AWS profiles.",
		Run:   c.ListProfilesRun,
	}

	return cmd
}

func (c *AwsCommands) ListProfilesRun(cmd *cobra.Command, args []string) {
	profileNames := c.Functions.GetAllProfileNames()

	for _, value := range profileNames {
		println("- " + value)
	}
}
