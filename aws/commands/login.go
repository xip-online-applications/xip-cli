package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

var loginProfile string

func Login() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to the SSO.",
		Run:   LoginRun,
	}

	cmd.Flags().StringVarP(&loginProfile, "profile", "p", "xip", "The profile to use")

	return cmd
}

func LoginRun(cmd *cobra.Command, args []string) {
	profile, _ := cmd.Flags().GetString("profile")

	path, _ := cmd.Flags().GetString("config")

	functions.Login(profile)
	functions.Sync(path, profile)
}
