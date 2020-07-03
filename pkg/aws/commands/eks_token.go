package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *AwsCommands) EksToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eks-token",
		Short: "Retrieve the EKS token to access kubernetes in it",
		Args:  cobra.ExactArgs(0),
		Run:   c.EksTokenRun,
	}

	defaultProfile, _ := c.Functions.GetDefaultProfile()

	cmd.Flags().String("cluster-name", "c", "The cluster name to get the token of")
	cmd.Flags().StringP("role-arn", "r", "", "The role to assume")
	cmd.Flags().StringP("profile", "p", defaultProfile, "The profile name to use")

	return cmd
}

func (c *AwsCommands) EksTokenRun(cmd *cobra.Command, args []string) {
	clusterName, _ := cmd.Flags().GetString("cluster-name")
	roleArn, _ := cmd.Flags().GetString("role-arn")
	profile, _ := cmd.Flags().GetString("profile")

	token, tokenExpiration, err := c.Functions.GetEksToken(profile, clusterName, roleArn)
	if err != nil {
		panic(err)
	}

	fmt.Println("{\"kind\": \"ExecCredential\",\"apiVersion\": \"client.authentication.k8s.io/v1alpha1\",\"spec\": {},\"status\": {\"expirationTimestamp\": \"" + tokenExpiration + "\",\"token\": \"" + token + "\"}}")
}
