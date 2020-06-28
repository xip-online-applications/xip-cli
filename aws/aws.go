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

	functs := functions.New()
	if functs == nil {
		return cmd
	}

	cmds := commands.New(*functs)

	cmd.AddCommand(cmds.Configure())
	cmd.AddCommand(cmds.Login())
	cmd.AddCommand(cmds.AddProfile())
	cmd.AddCommand(cmds.Default())
	cmd.AddCommand(cmds.Sync())
	cmd.AddCommand(cmds.Identity())

	return cmd
}
