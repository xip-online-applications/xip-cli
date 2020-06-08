package main

import (
	"github.com/spf13/cobra"
	"xip/aws"
	"xip/kubectl"
	"xip/utils"
)

func main() {
	cmd := &cobra.Command{
		Use: "xip",
	}

	cmd.AddCommand(aws.Aws())
	cmd.AddCommand(kubectl.Kubectl())
	cmd.AddCommand(utils.Version())

	cmd.Execute()
}
