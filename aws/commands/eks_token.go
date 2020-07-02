package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (c *AwsCommands) EksToken() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eks-token [profile name] [cluster name] [role arn]",
		Short: "Retrieve the EKS token to access kubernetes in it",
		Args:  cobra.RangeArgs(2, 3),
		Run:   c.EksTokenRun,
	}

	return cmd
}

func (c *AwsCommands) EksTokenRun(cmd *cobra.Command, args []string) {
	profile := args[0]
	clusterName := args[1]

	roleArn := ""
	if len(args) == 3 {
		roleArn = args[2]
	}

	token, tokenExpiration, err := c.Functions.GetEksToken(profile, clusterName, roleArn)
	if err != nil {
		panic(err)
	}

	fmt.Println("{\"kind\": \"ExecCredential\",\"apiVersion\": \"client.authentication.k8s.io/v1alpha1\",\"spec\": {},\"status\": {\"expirationTimestamp\": \"" + tokenExpiration + "\",\"token\": \"" + token + "\"}}")
}
