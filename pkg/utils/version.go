package utils

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Version() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Retrieve the version of the binary",
		Run:   VersionRun,
	}
}

func VersionRun(cmd *cobra.Command, args []string) {
	fmt.Println("Version: UNKNOWN")
}
