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

	cmd.Flags().BoolP("all", "a", false, "Run for all profiles")

	return cmd
}

func (c *AwsCommands) LoginRun(cmd *cobra.Command, args []string) {
	all, _ := cmd.Flags().GetBool("all")

	if all == true {
		currentDefault, _ := c.Functions.GetDefaultProfile()
		if currentDefault == nil {
			currentDefault = new(string)
		}

		for _, value := range c.Functions.GetAllSsoProfileNames() {
			c.Functions.Login(value)
		}

		c.Functions.SetDefault(*currentDefault)
	} else if len(args) == 1 {
		c.Functions.Login(args[0])
	} else {
		profile, err := c.Functions.GetDefaultProfile()
		if err != nil {
			panic("No default profile found.")
		}

		c.Functions.Login(*profile)
	}
}
