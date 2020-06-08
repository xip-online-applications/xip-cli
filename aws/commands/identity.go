package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Identity() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "identity",
		Short: "Retrieve current identity",
		Run:   IdentityRun,
	}

	return cmd
}

func IdentityRun(cmd *cobra.Command, args []string) {
	fmt.Println(functions.Identity())
}
