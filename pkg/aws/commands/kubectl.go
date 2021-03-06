package commands

import (
	"github.com/spf13/cobra"
)

func (c *AwsCommands) Kubectl() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kubectl [cluster name] [role ARN]",
		Short: "Configure the kubectl command tool for AWS.",
		Args:  cobra.RangeArgs(1, 2),
		Run:   c.KubectlRun,
	}

	defaultProfile, err := c.Functions.GetDefaultProfile()
	if err != nil {
		defaultProfile = "default"
	}

	cmd.Flags().StringP("kubectl-config-file", "k", "~/.kube/config", "The Kubectl config file")
	cmd.Flags().StringP("profile", "p", defaultProfile, "The profile name to use")
	cmd.Flags().StringP("namespace", "n", "", "The default namespace")
	cmd.Flags().StringP("alias", "a", "", "Alias name for this context")

	return cmd
}

func (c *AwsCommands) KubectlRun(cmd *cobra.Command, args []string) {
	clusterName := args[0]

	roleArn := ""
	if len(args) == 2 {
		roleArn = args[1]
	}

	profile, _ := cmd.Flags().GetString("profile")
	namespace, _ := cmd.Flags().GetString("namespace")
	alias, _ := cmd.Flags().GetString("alias")

	defaultUser, _ := c.Functions.GetDefaultProfile()
	defer c.Functions.SetDefault(defaultUser)

	if len(profile) > 0 {
		c.Functions.SetDefault(profile)
	}

	if err := c.Functions.RegisterKubectlProfile(clusterName, roleArn, profile, namespace, alias); err != nil {
		panic(err)
	}
}
