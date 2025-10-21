package main

import "testing"

func TestNewRootCmd(t *testing.T) {
	t.Helper()

	if cmd := newRootCmd(); cmd == nil {
		t.Fatal("newRootCmd returned nil")
	}
}
