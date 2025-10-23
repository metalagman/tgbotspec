package parser_test

import (
	"testing"

	"tgbotspec/internal/parser"
)

func TestTypeRefContainsType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		target string
		want   bool
	}{
		{name: "nil receiver", raw: "", target: "InputFile", want: false},
		{name: "empty target", raw: "InputFile", target: "  ", want: false},
		{name: "direct match", raw: "InputFile", target: "InputFile", want: true},
		{name: "case insensitive", raw: "inputfile", target: "InputFile", want: true},
		{name: "array nesting", raw: "Array of Array of InputFile", target: "InputFile", want: true},
		{name: "union contains", raw: "Message or InputFile", target: "InputFile", want: true},
		{name: "union no match", raw: "Message or Sticker", target: "InputFile", want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr *parser.TypeRef
			if tc.raw != "" {
				tr = parser.NewTypeRef(tc.raw)
			}

			if got := tr.ContainsType(tc.target); got != tc.want {
				t.Fatalf("ContainsType(%q) on %q = %v, want %v", tc.target, tc.raw, got, tc.want)
			}
		})
	}
}

func TestTypeRefContainsTypeWithPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		raw    string
		prefix string
		want   bool
	}{
		{name: "empty prefix", raw: "InputMediaPhoto", prefix: "", want: false},
		{name: "direct match", raw: "InputMediaPhoto", prefix: "InputMedia", want: true},
		{name: "array nesting", raw: "Array of InputMediaAnimation", prefix: "InputMedia", want: true},
		{name: "union match", raw: "String or InputStickerAnimated", prefix: "InputSticker", want: true},
		{name: "union no match", raw: "String or Message", prefix: "InputMedia", want: false},
		{name: "case insensitive", raw: "inputmediaphoto", prefix: "inputmedia", want: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var tr *parser.TypeRef
			if tc.raw != "" {
				tr = parser.NewTypeRef(tc.raw)
			}

			if got := tr.ContainsTypeWithPrefix(tc.prefix); got != tc.want {
				t.Fatalf("ContainsTypeWithPrefix(%q) on %q = %v, want %v", tc.prefix, tc.raw, got, tc.want)
			}
		})
	}
}
