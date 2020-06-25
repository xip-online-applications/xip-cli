package commands

import (
	"github.com/spf13/cobra"
	"xip/aws/functions"
)

func Kubectl() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl [cluster name] [role ARN]",
		Short: "Configure the kubectl command tool for AWS.",
		Args:  cobra.RangeArgs(1, 2),
		Run:   KubectlRun,
	}

	cmd.Flags().StringP("kubectl-config-file", "k", "~/.kube/config", "The Kubectl config file")
	cmd.Flags().StringP("profile", "p", functions.GetDefaultProfile(), "The profile name to use")
	cmd.Flags().StringP("namespace", "n", "", "The default namespace")
	cmd.Flags().StringP("alias", "a", "", "Alias name for this context")

	return cmd
}

func KubectlRun(cmd *cobra.Command, args []string) {
	clusterName := args[0]

	roleArn := ""
	if len(args) == 2 {
		roleArn = args[1]
	}

	awsConfigFile, _ := cmd.Flags().GetString("config")
	profile, _ := cmd.Flags().GetString("profile")
	namespace, _ := cmd.Flags().GetString("namespace")
	alias, _ := cmd.Flags().GetString("alias")

}
