package commands

import (
	"github.com/spf13/cobra"
)

func (c *AwsCommands) EcrLogin() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ecr-login [profile name]",
		Short: "Retrieve the ECR Docker login password",
		Args:  cobra.ExactArgs(0),
		Run:   c.EcrLoginRun,
	}

	cmd.Flags().StringP("profile", "p", "p", "The profile name to use")

	return cmd
}

func (c *AwsCommands) EcrLoginRun(cmd *cobra.Command, args []string) {
	profile, _ := cmd.Flags().GetString("profile")
	if len(profile) == 0 {
		profile, _ = c.Functions.GetDefaultProfile()
	}

	if len(profile) == 0 {
		panic("No profile provided and no default profile found.")
	}

	password, _ := c.Functions.GetEcrPassword(profile)
	print(password)
}
