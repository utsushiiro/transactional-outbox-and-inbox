package consumer

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "consumer",
		Short: "run consumer server",
		Run: func(cmd *cobra.Command, args []string) {
			run()
		},
	}

	return cmd
}
