package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func AddProfile() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add [profile name] [role arn] [source profile]",
		Short: "Add a new role by providing the Role ARN and possibly the source profile",
		Args:  cobra.RangeArgs(2, 3),
		Run:   AddProfileRun,
	}

	return cmd
}

func AddProfileRun(cmd *cobra.Command, args []string) {
	profile := args[0]
	role := args[1]

	sourceProfile := functions.GetDefaultProfile()
	if len(args) == 3 {
		sourceProfile = args[2]
	}

	path, _ := cmd.Flags().GetString("config")

	functions.CreateOrUpdateRoleAssumeProfile(path, profile, sourceProfile, role)
	functions.SetDefault(path, profile)
}
