package parser

import (
	"errors"
	"strings"
)

// ErrElementNotFound signals that the expected anchor/header could not be
// located within the Telegram documentation.
var ErrElementNotFound = errors.New("element not found")

// ErrReturnTypeNotParsed indicates we failed to recognise a return type in the
// method description block.
var ErrReturnTypeNotParsed = errors.New("method return type not parsed")

// isOptionalDescription returns true if the description starts with the word
// "Optional" (case-insensitive), ignoring leading spaces and allowing
// punctuation like a trailing dot right after the word.
func isOptionalDescription(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	ls := strings.ToLower(s)
	if strings.HasPrefix(ls, "optional") {
		return true
	}
	return false
}
