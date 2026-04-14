package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPeratioCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "peratio",
		Short: "P/E ratio series from MacroTrends (placeholder)",
		Long:  `Fetch price-to-earnings ratio history from MacroTrends. Not implemented yet.`,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return fmt.Errorf("peratio: not implemented yet (placeholder subcommand)")
		},
	}
}
