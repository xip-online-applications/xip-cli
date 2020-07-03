package commands

import "xip/aws/functions"

type AwsCommands struct {
	Functions *functions.Functions
}

func New(f functions.Functions) *AwsCommands {
	return &AwsCommands{
		Functions: &f,
	}
}
