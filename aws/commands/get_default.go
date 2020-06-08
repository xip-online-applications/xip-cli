package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func GetDefault() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-default",
		Short: "Retrieve the current default profile",
		Run:   GetDefaultRun,
	}

	return cmd
}

func GetDefaultRun(cmd *cobra.Command, args []string) {
	fmt.Println(functions.GetDefault())
}
