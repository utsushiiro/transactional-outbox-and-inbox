package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/utsushiiro/transactional-outbox-and-inbox/cmd/consumer"
	"github.com/utsushiiro/transactional-outbox-and-inbox/cmd/producer"
)

func main() {
	if err := newCmdRoot().Execute(); err != nil {
		os.Exit(1)
	}
}

func newCmdRoot() *cobra.Command {
	var rootCmd = &cobra.Command{
		Use:   "server",
		Short: "producer/consumer server",
	}

	rootCmd.AddCommand(producer.NewCmd())
	rootCmd.AddCommand(consumer.NewCmd())

	return rootCmd
}
