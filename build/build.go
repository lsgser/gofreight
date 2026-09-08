package build

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// Options configures a production build.
type Options struct {
	Dir        string
	Output     string
	GOOS       string
	GOARCH     string
	LDFlags    string
	CGOEnabled string
	RouteCache bool
	TrimPath   bool
	Verbose    bool
}

// Result describes a completed build.
type Result struct {
	Output string
	GOOS   string
	GOARCH string
}

// DefaultOutput returns bin/<module-base> from go.mod in dir.
func DefaultOutput(dir string) (string, error) {
	mod, err := ModuleName(dir)
	if err != nil {
		return "", err
	}
	base := mod
	if i := strings.LastIndex(base, "/"); i >= 0 {
		base = base[i+1:]
	}
	return filepath.Join("bin", base), nil
}

// ModuleName reads the module path from go.mod.
func ModuleName(dir string) (string, error) {
	path := filepath.Join(dir, "go.mod")
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("go.mod not found — run from application root")
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	return "", fmt.Errorf("module directive not found in go.mod")
}

// ValidateAppRoot returns an error when dir is not a Go application root.
func ValidateAppRoot(dir string) error {
	if _, err := os.Stat(filepath.Join(dir, "main.go")); err != nil {
		return fmt.Errorf("main.go not found — run gofreight build from your application root")
	}
	if _, err := os.Stat(filepath.Join(dir, "go.mod")); err != nil {
		return fmt.Errorf("go.mod not found — run gofreight build from your application root")
	}
	return nil
}

// RunRouteCache runs the app once to cache routes (optional pre-build step).
func RunRouteCache(dir string) error {
	cmd := exec.Command("go", "run", ".")
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GOFREIGHT_ROUTE_CACHE=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Build compiles the application binary for production.
func Build(opts Options) (Result, error) {
	dir := opts.Dir
	if dir == "" {
		dir = "."
	}
	if err := ValidateAppRoot(dir); err != nil {
		return Result{}, err
	}

	output := opts.Output
	if output == "" {
		var err error
		output, err = DefaultOutput(dir)
		if err != nil {
			return Result{}, err
		}
	}

	if err := os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return Result{}, err
	}

	if opts.RouteCache {
		if err := RunRouteCache(dir); err != nil {
			return Result{}, fmt.Errorf("route cache failed (use --no-route-cache to skip): %w", err)
		}
	}

	goos := opts.GOOS
	if goos == "" {
		goos = os.Getenv("GOOS")
	}
	if goos == "" {
		goos = runtime.GOOS
	}
	goarch := opts.GOARCH
	if goarch == "" {
		goarch = os.Getenv("GOARCH")
	}
	if goarch == "" {
		goarch = runtime.GOARCH
	}

	ldflags := opts.LDFlags
	if ldflags == "" {
		ldflags = "-s -w"
	}

	cgo := opts.CGOEnabled
	if cgo == "" {
		cgo = "0"
	}

	args := []string{"build", "-ldflags=" + ldflags, "-o", output, "."}
	if opts.TrimPath {
		args = []string{"build", "-trimpath", "-ldflags=" + ldflags, "-o", output, "."}
	}

	cmd := exec.Command("go", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"CGO_ENABLED="+cgo,
		"GOOS="+goos,
		"GOARCH="+goarch,
	)
	if opts.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		return Result{}, fmt.Errorf("go build failed: %w", err)
	}

	return Result{Output: output, GOOS: goos, GOARCH: goarch}, nil
}

// DeployArtifacts lists paths to copy alongside the binary for production.
func DeployArtifacts() []string {
	return []string{
		".env",
		"app/views/",
		"public/",
		"db/",
		"config/",
		"storage/",
	}
}
