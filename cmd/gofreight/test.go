package main

/*
|--------------------------------------------------------------------------
| Test
|--------------------------------------------------------------------------
|
| Implements Test as part of the gofreight package in the Gofreight
| framework.
| 
| This directory contains the gofreight CLI binary: command registration,
| terminal UI, and handlers for make:*, migrate, serve, test, and
| mail:preview.
| 
| Each subcommand lives in its own source file; commands.go registers the
| catalog shown by gofreight list.
| 
| Install locally with go install ./cmd/gofreight from the framework
| repository root.
| 
*/

import (
	"fmt"
	"os"

	"github.com/lsgser/gofreight/testrunner"
)

func registerTestCommands() {
	register(command{
		Name:        "test",
		Category:    "app",
		Description: "Run Gofreight feature tests (gftest in tests/)",
		Usage:       "gofreight test [flags] [packages]",
		Run:         handleFeatureTest,
	})
	register(command{
		Name:        "test:unit",
		Category:    "app",
		Description: "Run Go unit tests (app packages)",
		Usage:       "gofreight test:unit [flags] [packages]",
		Run:         handleUnitTest,
	})
}

func handleFeatureTest(args []string) {
	runTests(testrunner.ModeFeature, args)
}

func handleUnitTest(args []string) {
	runTests(testrunner.ModeUnit, args)
}

func runTests(mode testrunner.Mode, args []string) {
	opts, packages, help := testrunner.ParseArgs(args)
	if help {
		printTestHelp(mode)
		return
	}
	opts.Mode = mode
	opts.Packages = packages

	if err := testrunner.ValidateFeatureTestsDir(mode); err != nil {
		fail(err)
	}

	fmt.Printf("Running %s (GOFREIGHT_ENV=test)\n", testrunner.SuiteLabel(mode))
	if len(opts.Packages) == 0 {
		fmt.Printf("  packages: %v\n\n", testrunner.DefaultPackages(mode))
	} else {
		fmt.Printf("  packages: %v\n\n", opts.Packages)
	}

	if err := testrunner.Run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func printTestHelp(mode testrunner.Mode) {
	switch mode {
	case testrunner.ModeUnit:
		fmt.Println(`Run Go unit tests for application code (models, services, etc.).

Usage:
  gofreight test:unit [flags] [packages]

Defaults:
  ./app/... when app/ exists, otherwise ./...

Flags:
  -v, --verbose       Verbose test output
      --cover         Enable coverage
      --filter, -run  Run tests matching a regex
      --count N       Run each test N times (default 1)

Examples:
  gofreight test:unit
  gofreight test:unit ./app/models/...
  gofreight test:unit -v --filter TestUser

For raw Go test control, use: go test ./...`)
	default:
		fmt.Println(`Run Gofreight feature tests (HTTP, gftest, factories) from tests/.

Usage:
  gofreight test [flags] [packages]

Defaults:
  ./tests/... when tests/ exists

Flags:
  -v, --verbose       Verbose test output
      --cover         Enable coverage
      --filter, -run  Run tests matching a regex
      --count N       Run each test N times (default 1)

Examples:
  gofreight test
  gofreight test -v --filter TestPosts
  gofreight test ./tests/...

Generate tests: gofreight make:test Posts
Unit tests:     gofreight test:unit`)
	}
}
