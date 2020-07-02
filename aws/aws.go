package aws

import (
	"github.com/spf13/cobra"

	"xip/aws/commands"
	"xip/aws/functions"
)

func Aws() *cobra.Command {
	cmd := &cobra.Command{
		Use: "aws",
	}

	cmds := commands.New(functions.New())

	cmd.AddCommand(cmds.Configure())
	cmd.AddCommand(cmds.Login())
	cmd.AddCommand(cmds.AddProfile())
	cmd.AddCommand(cmds.Default())
	cmd.AddCommand(cmds.Kubectl())
	cmd.AddCommand(cmds.Identity())
	cmd.AddCommand(cmds.EksToken())

	return cmd
}
