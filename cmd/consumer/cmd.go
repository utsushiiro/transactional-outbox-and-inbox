package consumer

import (
	"github.com/spf13/cobra"
)

func NewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "consumer",
		Short: "run consumer server",
		Run: func(_ *cobra.Command, _ []string) {
			run()
		},
	}

	return cmd
}
