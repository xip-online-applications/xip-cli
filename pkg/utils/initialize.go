package utils

import (
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"

	"xip/aws/functions/config/app"
)

func Initialize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the X-IP tool!",
		Run:   InitializeRun,
	}

	usr, _ := user.Current()
	cmd.Flags().StringP("config", "c", filepath.FromSlash(usr.HomeDir+"/.aws/config"), "AWS config file path")

	return cmd
}

func InitializeRun(cmd *cobra.Command, args []string) {
	appConfig := app.New()
	values := appConfig.Get()

	// Update app configuration
	appConfig.Set(values)
}
