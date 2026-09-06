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

func TestPrintUsageDoesNotDuplicateBanner(t *testing.T) {
	initCommands()
	var buf bytes.Buffer
	printCLIBanner(&buf)
	printCommandListTo(&buf, "", false)
	out := buf.String()
	if strings.Count(out, "Batteries-included web framework for Go") != 1 {
		t.Fatalf("expected banner once, got:\n%s", out)
	}
	if !strings.Contains(out, "Available commands") {
		t.Fatalf("expected command list in output:\n%s", out)
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
