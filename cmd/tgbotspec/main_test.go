package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/metalagman/tgbotspec/internal/scraper"
)

func TestNewRootCmd(t *testing.T) {
	originalRun := runScraper
	runScraper = func(w io.Writer, opts scraper.Options) error {
		if w == nil {
			t.Fatal("expected writer to be non-nil")
		}

		_, _ = w.Write([]byte("ok"))

		return nil
	}

	t.Cleanup(func() {
		runScraper = originalRun
	})

	cmd := newRootCmd()
	if cmd.Use != "tgbotspec" {
		t.Fatalf("unexpected Use: %q", cmd.Use)
	}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if buf.String() != "ok" {
		t.Fatalf("expected RunE to write to output, got %q", buf.String())
	}
}

func TestNewRootCmdOutputFlag(t *testing.T) {
	originalRun := runScraper
	runScraper = func(w io.Writer, opts scraper.Options) error {
		if w == nil {
			t.Fatal("expected writer to be non-nil")
		}

		_, _ = w.Write([]byte("ok"))

		return nil
	}

	t.Cleanup(func() {
		runScraper = originalRun
	})

	cmd := newRootCmd()
	outputDir := t.TempDir()
	outputPath := outputDir + string(os.PathSeparator) + "spec.yaml"

	if err := os.WriteFile(outputPath, []byte("old"), 0o600); err != nil {
		t.Fatalf("failed to write output file: %v", err)
	}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"-o", outputPath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}

	if buf.Len() != 0 {
		t.Fatalf("expected stdout to be empty, got %q", buf.String())
	}

	contents, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read output file: %v", err)
	}

	if string(contents) != "ok" {
		t.Fatalf("expected output file to be overwritten, got %q", string(contents))
	}
}

func TestNewRootCmdError(t *testing.T) {
	cmd := newRootCmd()
	// Use an invalid path to trigger os.Create error
	cmd.SetArgs([]string{"-o", "/invalid/path/spec.yaml"})
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)

	if err := cmd.Execute(); err == nil {
		t.Fatal("expected Execute to return error for invalid output path")
	}
}

func TestMainSuccess(t *testing.T) {
	originalRun := runScraper
	originalExit := exit
	originalArgs := os.Args

	runScraper = func(w io.Writer, opts scraper.Options) error {
		return nil
	}

	exitCalled := false
	exit = func(code int) {
		exitCalled = true
	}

	os.Args = []string{"tgbotspec"}

	t.Cleanup(func() {
		runScraper = originalRun
		exit = originalExit
		os.Args = originalArgs
	})

	main()

	if exitCalled {
		t.Fatal("exit should not have been called on successful execution")
	}
}

func TestMainExitOnError(t *testing.T) {
	originalRun := runScraper
	originalExit := exit
	originalArgs := os.Args

	runScraper = func(w io.Writer, opts scraper.Options) error {
		return errors.New("boom")
	}

	var (
		exitCalled bool
		exitCode   int
	)

	exit = func(code int) {
		exitCalled = true
		exitCode = code
	}

	os.Args = []string{"tgbotspec"}

	t.Cleanup(func() {
		runScraper = originalRun
		exit = originalExit
		os.Args = originalArgs
	})

	main()

	if !exitCalled {
		t.Fatal("expected exit to be called when command fails")
	}

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
}
