package commands

import (
	"github.com/spf13/cobra"

	"xip/aws/functions/sso"
)

func (c *AwsCommands) Configure() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configure [role] [sso start url] [account id]",
		Short: "Configure AWS CLI SSO by providing the role. This role will be provided to you by your administrator.",
		Args:  cobra.ExactArgs(3),
		Run:   c.ConfigureRun,
	}

	cmd.Flags().StringP("profile", "p", "p", "The profile name to use")
	cmd.Flags().StringP("region", "r", "eu-west-1", "The region of your org")

	return cmd
}

func (c *AwsCommands) ConfigureRun(cmd *cobra.Command, args []string) {
	role := args[0]
	startUrl := args[1]
	accountId := args[2]

	profile, _ := cmd.Flags().GetString("profile")
	region, _ := cmd.Flags().GetString("region")

	c.Functions.Configure(sso.ConfigureValues{
		Region:    &region,
		StartUrl:  &startUrl,
		Profile:   &profile,
		AccountId: &accountId,
		RoleName:  &role,
	})
	c.Functions.Login(profile)
	c.Functions.PrintDefaultHelp(profile)
}
