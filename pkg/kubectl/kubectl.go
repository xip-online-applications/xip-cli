package kubectl

import "github.com/spf13/cobra"

func Kubectl() *cobra.Command {
	cmd := &cobra.Command{Use: "kubectl"}

	cmd.AddCommand(Test())

	return cmd
}
