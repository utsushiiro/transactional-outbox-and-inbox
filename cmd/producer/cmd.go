package producer

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "producer",
		Short: "run producer server",
		Run: func(_ *cobra.Command, _ []string) {
			run()
		},
	}

	return cmd
}
