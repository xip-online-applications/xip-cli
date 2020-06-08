package aws

import (
	"github.com/spf13/cobra"
	"os/user"
	"xip/aws/commands"
)

var configFilePath string

func Aws() *cobra.Command {
	cmd := &cobra.Command{
		Use: "aws",
	}

	cmd.AddCommand(commands.Configure())
	cmd.AddCommand(commands.Default())
	cmd.AddCommand(commands.Login())
	cmd.AddCommand(commands.AddProfile())
	cmd.AddCommand(commands.GetDefault())
	cmd.AddCommand(commands.Sync())

	usr, _ := user.Current()
	cmd.PersistentFlags().StringVarP(&configFilePath, "config", "c", usr.HomeDir+"/.aws/config", "AWS config file path")

	return cmd
}
