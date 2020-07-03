package commands

import (
	"github.com/spf13/cobra"
)

func (c *AwsCommands) Login() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [profile]",
		Short: "Login to the SSO. If profile is omitted, the current default will be logged in.",
		Args:  cobra.RangeArgs(0, 1),
		Run:   c.LoginRun,
	}

	return cmd
}

func (c *AwsCommands) LoginRun(cmd *cobra.Command, args []string) {
	if len(args) == 1 {
		profile := args[0]

		c.Functions.Login(profile)
		c.Functions.PrintDefaultHelp(profile)
	} else {
		c.Functions.Login("")
	}
}
