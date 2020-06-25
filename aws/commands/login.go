package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Login() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login [profile]",
		Short: "Login to the SSO.",
		Args:  cobra.RangeArgs(0, 1),
		Run:   LoginRun,
	}

	return cmd
}

func LoginRun(cmd *cobra.Command, args []string) {
	path, _ := cmd.Flags().GetString("config")

	if len(args) == 1 {
		functions.Login(args[0])
	} else {
		for _, value := range functions.GetAllSsoProfileNames(path) {
			functions.Login(value)
		}
	}

	for _, value := range functions.GetAllProfileNames(path) {
		functions.Sync(path, value)
	}
}
