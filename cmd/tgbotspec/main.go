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
	var outputPath string

	cmd := &cobra.Command{
		Use:          "tgbotspec",
		Short:        "Generate an OpenAPI spec for the Telegram Bot API",
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			output := cmd.OutOrStdout()

			if outputPath != "" {
				file, err := os.Create(outputPath)
				if err != nil {
					return fmt.Errorf("create output file: %w", err)
				}

				defer func() {
					if closeErr := file.Close(); closeErr != nil && err == nil {
						err = fmt.Errorf("close output file: %w", closeErr)
					}
				}()

				output = file
			}

			if err := runScraper(output); err != nil {
				return fmt.Errorf("run scraper: %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "Write output to file instead of stdout")

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
