package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func AddProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile name] [role arn]",
		Short: "Add a new role by providing the Role ARN",
		Args:  cobra.ExactArgs(2),
		Run:   AddProfileRun,
	}

	cmd.Flags().StringP("source-profile", "p", functions.GetDefaultProfile(), "The source profile name to use")

	return cmd
}

func AddProfileRun(cmd *cobra.Command, args []string) {
	role := args[1]
	profile := args[0]

	path, _ := cmd.Flags().GetString("config")
	sourceProfile, _ := cmd.Flags().GetString("profile")

	functions.CreateOrUpdateRoleAssumeProfile(path, profile, sourceProfile, role)
	functions.SetDefault(path, profile)
}
