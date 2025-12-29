package main

import (
	"fmt"
	"os"

	"github.com/metalagman/tgbotspec/internal/scraper"

	"github.com/spf13/cobra"
)

var (
	runScraper = scraper.Run
	exit       = os.Exit
)

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tgbotspec",
		Short:        "Generate an OpenAPI spec for the Telegram Bot API",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := runScraper(cmd.OutOrStdout()); err != nil {
				return fmt.Errorf("run scraper: %w", err)
			}

			return nil
		},
	}

	return cmd
}

func execute() error {
	rootCmd := newRootCmd()
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)

	return rootCmd.Execute()
}

func main() {
	if err := execute(); err != nil {
		exit(1)
	}
}
