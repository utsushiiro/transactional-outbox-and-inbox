package producer

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "producer",
		Short: "run producer server",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	return cmd
}
