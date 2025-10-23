package main

import (
	"bytes"
	"errors"
	"io"
	"os"
	"testing"
)

func TestNewRootCmd(t *testing.T) {
	originalRun := runScraper
	runScraper = func(w io.Writer) error {
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

func TestMainSuccess(t *testing.T) {
	originalRun := runScraper
	originalExit := exit
	originalArgs := os.Args

	runScraper = func(w io.Writer) error {
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

	runScraper = func(w io.Writer) error {
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
