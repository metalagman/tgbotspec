package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"tgbotspec/internal/scraper"
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tgbotspec",
		Short:        "Generate an OpenAPI spec for the Telegram Bot API",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := scraper.Run(cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("run scraper: %w", err)
			}
			return nil
		},
	}
	return cmd
}

func main() {
	rootCmd := newRootCmd()
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
