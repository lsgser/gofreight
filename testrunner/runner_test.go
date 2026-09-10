package testrunner_test

/*
|--------------------------------------------------------------------------
| Runner
|--------------------------------------------------------------------------
|
| Test suite for Runner in the testrunner package.
| 
| Uses table-driven tests, httptest, or gftest where applicable. Failures
| should indicate regressions in public API or HTTP behavior.
| 
| testrunner powers gofreight test (feature tests in tests/) and gofreight
| test:unit (Go tests in app/).
| 
| It sets GOFREIGHT_ENV=test and forwards flags to go test with sensible
| default package paths.
| 
| Run with go test ./testrunner/... or go test for this package from the
| framework root.
| 
*/

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
