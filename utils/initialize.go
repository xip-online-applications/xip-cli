package utils

import (
	"github.com/spf13/cobra"
	"os/user"
	"xip/aws/functions/config/app"
)

func Initialize() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the X-IP tool!",
		Run:   InitializeRun,
	}

	usr, _ := user.Current()
	cmd.Flags().StringP("config", "c", usr.HomeDir+"/.aws/config", "AWS config file path")

	return cmd
}

func InitializeRun(cmd *cobra.Command, args []string) {
	appConfig := app.New()
	values := appConfig.Get()

	// Retrieve flag values
	awsConfigFilePath, _ := cmd.Flags().GetString("config")

	// SetSsoProfile the values
	values.AwsConfigPath = &awsConfigFilePath

	// Update app configuration
	appConfig.Set(values)
}
