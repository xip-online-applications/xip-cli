package commands

import (
	"github.com/spf13/cobra"
)

func (c *AwsCommands) AddProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile name] [role arn] [source profile]",
		Short: "Add a new role by providing the Role ARN and possibly the source profile",
		Args:  cobra.RangeArgs(2, 3),
		Run:   c.AddProfileRun,
	}

	return cmd
}

func (c *AwsCommands) AddProfileRun(cmd *cobra.Command, args []string) {
	profile := args[0]
	role := args[1]

	sourceProfile, _ := c.Functions.GetDefaultProfile()
	if len(args) == 3 {
		sourceProfile = args[2]
	}

	if len(sourceProfile) == 0 {
		panic("Source profile is empty.")
	}

	c.Functions.AddProfile(profile, sourceProfile, role)
}
