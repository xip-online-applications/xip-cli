package kubectl

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Test() *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Kubectl test",
		Run:   TestExecute,
	}
}

func TestExecute(cmd *cobra.Command, args []string) {
	fmt.Println("kubectl test")
}
