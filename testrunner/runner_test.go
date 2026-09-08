package testrunner_test

import (
	"testing"

	"github.com/lsgser/gofreight/testrunner"
)

func TestDefaultPackagesFeature(t *testing.T) {
	pkgs := testrunner.DefaultPackages(testrunner.ModeFeature)
	if len(pkgs) == 0 {
		t.Fatal("expected packages")
	}
}

func TestParseArgs(t *testing.T) {
	opts, pkgs, help := testrunner.ParseArgs([]string{"-v", "--filter", "TestHome", "./tests/..."})
	if help {
		t.Fatal("unexpected help")
	}
	if !opts.Verbose || opts.Run != "TestHome" {
		t.Fatalf("opts: %+v", opts)
	}
	if len(pkgs) != 1 || pkgs[0] != "./tests/..." {
		t.Fatalf("pkgs: %v", pkgs)
	}
}

func TestSuiteLabel(t *testing.T) {
	if testrunner.SuiteLabel(testrunner.ModeFeature) == "" {
		t.Fatal("expected feature label")
	}
	if testrunner.SuiteLabel(testrunner.ModeUnit) == "" {
		t.Fatal("expected unit label")
	}
}
