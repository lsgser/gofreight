package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestPrintCLIBanner(t *testing.T) {
	var buf bytes.Buffer
	printCLIBanner(&buf)
	out := buf.String()
	if !strings.Contains(out, "Gofreight") || !strings.Contains(out, "Batteries-included") {
		t.Fatalf("unexpected banner: %s", out)
	}
}

func TestPrintServeWelcome(t *testing.T) {
	var buf bytes.Buffer
	printServeWelcomeTo(&buf, 5000, "development")
	out := buf.String()
	for _, part := range []string{"Gofreight development server", "http://localhost:5000", "admin"} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected %q in output\n%s", part, out)
		}
	}
}

func TestPrintNewAppWelcomeIncludesAppName(t *testing.T) {
	var buf bytes.Buffer
	printNewAppWelcomeTo(&buf, "myblog")

	out := buf.String()
	for _, part := range []string{"Gofreight", "myblog", "gofreight key:generate", "http://localhost:5000", "What's next?"} {
		if !strings.Contains(out, part) {
			t.Fatalf("expected output to contain %q\n%s", part, out)
		}
	}
}
