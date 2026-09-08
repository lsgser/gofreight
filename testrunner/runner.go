package testrunner

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// Mode selects which test suite to run.
type Mode string

const (
	ModeFeature Mode = "feature" // gftest HTTP/feature tests in tests/
	ModeUnit    Mode = "unit"    // Go unit tests under app/
)

// Options configures a test run.
type Options struct {
	Mode     Mode
	Packages []string
	Run      string
	Verbose  bool
	Cover    bool
	Count    int
}

// DefaultPackages returns package paths for the given mode when none are specified.
func DefaultPackages(mode Mode) []string {
	switch mode {
	case ModeUnit:
		if dirExists("app") {
			return []string{"./app/..."}
		}
		return []string{"./..."}
	default:
		if dirExists("tests") {
			return []string{"./tests/..."}
		}
		return []string{"./..."}
	}
}

// Run executes go test with Gofreight test defaults.
func Run(opts Options) error {
	pkgs := opts.Packages
	if len(pkgs) == 0 {
		pkgs = DefaultPackages(opts.Mode)
	}

	args := []string{"test"}
	args = append(args, pkgs...)
	if opts.Verbose {
		args = append(args, "-v")
	}
	if opts.Cover {
		args = append(args, "-cover")
	}
	if opts.Run != "" {
		args = append(args, "-run", opts.Run)
	}
	if opts.Count > 0 {
		args = append(args, fmt.Sprintf("-count=%d", opts.Count))
	}

	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), "GOFREIGHT_ENV=test")

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("go test failed: %w", err)
	}
	return nil
}

// ParseArgs splits CLI flags from package paths.
func ParseArgs(args []string) (opts Options, packages []string, help bool) {
	opts = Options{Mode: ModeFeature, Count: 1}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "-h", arg == "--help":
			return opts, nil, true
		case arg == "-v", arg == "--verbose":
			opts.Verbose = true
		case arg == "--cover":
			opts.Cover = true
		case arg == "--filter", arg == "-run":
			if i+1 >= len(args) {
				return opts, nil, false
			}
			i++
			opts.Run = args[i]
		case strings.HasPrefix(arg, "--filter="):
			opts.Run = strings.TrimPrefix(arg, "--filter=")
		case strings.HasPrefix(arg, "-run="):
			opts.Run = strings.TrimPrefix(arg, "-run=")
		case arg == "--count":
			if i+1 >= len(args) {
				return opts, nil, false
			}
			i++
			fmt.Sscanf(args[i], "%d", &opts.Count)
		case strings.HasPrefix(arg, "--count="):
			fmt.Sscanf(strings.TrimPrefix(arg, "--count="), "%d", &opts.Count)
		case strings.HasPrefix(arg, "-"):
			// pass unknown flags through to go test
			packages = append(packages, arg)
		default:
			packages = append(packages, arg)
		}
	}
	return opts, packages, false
}

func dirExists(name string) bool {
	info, err := os.Stat(name)
	return err == nil && info.IsDir()
}

// SuiteLabel returns a human-readable suite name.
func SuiteLabel(mode Mode) string {
	switch mode {
	case ModeUnit:
		return "Go unit tests"
	default:
		return "Gofreight feature tests"
	}
}

// ValidateFeatureTestsDir errors when tests/ is missing for feature mode.
func ValidateFeatureTestsDir(mode Mode) error {
	if mode != ModeFeature {
		return nil
	}
	if !dirExists("tests") {
		return fmt.Errorf("tests/ directory not found — create feature tests with gofreight make:test <Name>")
	}
	return nil
}
