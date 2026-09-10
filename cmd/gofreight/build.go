package main

/*
|--------------------------------------------------------------------------
| Build
|--------------------------------------------------------------------------
|
| Implements Build as part of the gofreight package in the Gofreight
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
	"strings"

	"github.com/lsgser/gofreight/build"
)

func registerBuildCommand() {
	register(command{
		Name:        "build",
		Category:    "app",
		Description: "Compile the application for production deployment",
		Usage:       "gofreight build [-o bin/app] [--os linux] [--arch amd64] [--route-cache]",
		Run:         handleBuild,
	})
}

func handleBuild(args []string) {
	opts := build.Options{
		TrimPath: true,
		Verbose:  true,
	}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-o" || arg == "--output":
			if i+1 >= len(args) {
				failMsg("missing value for " + arg)
			}
			i++
			opts.Output = args[i]
		case strings.HasPrefix(arg, "-o="):
			opts.Output = strings.TrimPrefix(arg, "-o=")
		case strings.HasPrefix(arg, "--output="):
			opts.Output = strings.TrimPrefix(arg, "--output=")
		case arg == "--os":
			if i+1 >= len(args) {
				failMsg("missing value for --os")
			}
			i++
			opts.GOOS = args[i]
		case strings.HasPrefix(arg, "--os="):
			opts.GOOS = strings.TrimPrefix(arg, "--os=")
		case arg == "--arch":
			if i+1 >= len(args) {
				failMsg("missing value for --arch")
			}
			i++
			opts.GOARCH = args[i]
		case strings.HasPrefix(arg, "--arch="):
			opts.GOARCH = strings.TrimPrefix(arg, "--arch=")
		case arg == "--route-cache":
			opts.RouteCache = true
		case arg == "--no-route-cache":
			opts.RouteCache = false
		case arg == "--no-trimpath":
			opts.TrimPath = false
		case arg == "-h", arg == "--help":
			printBuildHelp()
			return
		default:
			failMsg("unknown flag: " + arg)
		}
	}

	result, err := build.Build(opts)
	if err != nil {
		fail(err)
	}

	fmt.Printf("\n  Production build complete\n")
	fmt.Printf("  Binary:   %s\n", result.Output)
	fmt.Printf("  Platform: %s/%s\n\n", result.GOOS, result.GOARCH)
	fmt.Println("  Deploy with:")
	for _, path := range build.DeployArtifacts() {
		fmt.Printf("    • %s\n", path)
	}
	fmt.Println("\n  On the server:")
	fmt.Println("    gofreight migrate --force")
	fmt.Printf("    GOFREIGHT_ENV=production ./%s\n\n", result.Output)
}

func printBuildHelp() {
	fmt.Println(`Compile your Gofreight app into a single production binary.

Usage:
  gofreight build [flags]

Flags:
  -o, --output <path>   Output binary path (default: bin/<module-name>)
      --os <goos>       Target GOOS (e.g. linux, darwin, windows)
      --arch <goarch>   Target GOARCH (e.g. amd64, arm64)
      --route-cache     Cache routes before building (requires runnable app)
      --no-route-cache  Skip route caching (default)
      --no-trimpath     Disable -trimpath for debuggable paths in binary

Examples:
  gofreight build
  gofreight build -o bin/myapp
  GOOS=linux GOARCH=amd64 gofreight build
  gofreight build --os linux --arch arm64

See docs/deployment.md for the full production checklist.`)
}

func failMsg(msg string) {
	fmt.Fprintln(os.Stderr, "Error:", msg)
	os.Exit(1)
}
