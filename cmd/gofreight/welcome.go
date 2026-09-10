package main

/*
|--------------------------------------------------------------------------
| Welcome
|--------------------------------------------------------------------------
|
| Implements Welcome as part of the gofreight package in the Gofreight
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
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"github.com/lsgser/gofreight/version"
)

func printCLIBanner(w io.Writer) {
	useColor := isTerminal(os.Stdout)
	c := newConsoleStyle(useColor)
	fmt.Fprintf(w, "\n%s\n", c.bold("  Gofreight")+c.dim("  v"+version.Version))
	fmt.Fprintf(w, "%s\n\n", c.dim("  Batteries-included web framework for Go"))
}

func printServeWelcome(port int, env string) {
	printServeWelcomeTo(os.Stdout, port, env)
}

func printServeWelcomeTo(w io.Writer, port int, env string) {
	useColor := isTerminal(os.Stdout)
	c := newConsoleStyle(useColor)
	url := fmt.Sprintf("http://localhost:%d", port)
	admin := fmt.Sprintf("http://localhost:%d/admin", port)

	fmt.Fprintf(w, "\n")
	fmt.Fprintf(w, "%s\n", c.cyan("  ╭──────────────────────────────────────────────────────────────╮"))
	fmt.Fprintf(w, "%s\n", line(c, "  Gofreight development server"))
	fmt.Fprintf(w, "%s\n", line(c, ""))
	fmt.Fprintf(w, "%s\n", line(c, "  Local   "+url))
	if env == "development" || env == "local" || env == "" {
		fmt.Fprintf(w, "%s\n", line(c, "  Admin   "+admin))
	}
	fmt.Fprintf(w, "%s\n", c.cyan("  ╰──────────────────────────────────────────────────────────────╯"))
	fmt.Fprintf(w, "\n%s\n\n", c.dim("  Press Ctrl+C to stop the server."))
}

func printNewAppWelcome(appName string) {
	printNewAppWelcomeTo(os.Stdout, appName)
}

func printNewAppWelcomeTo(w io.Writer, appName string) {
	useColor := isTerminal(os.Stdout)
	c := newConsoleStyle(useColor)

	write := func(format string, args ...any) {
		fmt.Fprintf(w, format, args...)
	}

	write("\n")
	write("%s\n", c.cyan("  ╭──────────────────────────────────────────────────────────────╮"))
	write("%s\n", line(c, "  Gofreight  v"+version.Version))
	write("%s\n", line(c, "  Batteries-included web framework for Go"))
	write("%s\n", line(c, ""))
	write("%s\n", line(c, "  ✓  Application ready: "+appName))
	write("%s\n", c.cyan("  ╰──────────────────────────────────────────────────────────────╯"))
	write("\n")
	write("%s\n\n", c.bold("  What's next?"))
	write("    %s cd %s\n", c.dim("1."), appName)
	write("    %s go mod tidy\n", c.dim("2."))
	write("    %s gofreight key:generate\n", c.dim("3."))
	write("    %s gofreight db:create\n", c.dim("4."))
	write("    %s gofreight migrate\n", c.dim("5."))
	write("    %s gofreight serve\n", c.dim("6."))
	write("\n")
	write("%s\n\n", c.bold("  When the server is running"))
	write("    App     %s\n", c.cyan("http://localhost:5000"))
	write("    Admin   %s\n", c.dim("http://localhost:5000/admin  (development only)"))
	write("\n")
	write("%s\n\n", c.bold("  Learn more"))
	write("    gofreight list              %s\n", c.dim("all CLI commands"))
	write("    gofreight make:scaffold     %s\n", c.dim("generate models, views, routes"))
	write("    %s\n", c.cyan("https://lsgser.github.io/gofreight-web/"))
	write("                              %s\n", c.dim("documentation & tutorials"))
	write("\n")
	write("  %s\n\n", c.dim("Happy shipping with Gofreight."))
}

func line(c consoleStyle, text string) string {
	inner := 62
	if text == "" {
		return c.cyan("  │") + strings.Repeat(" ", inner) + c.cyan("│")
	}
	marker, body := "", text
	if strings.HasPrefix(text, "  ✓  ") {
		marker = c.green("  ✓")
		body = text[4:]
	}
	content := marker + body
	if marker != "" {
		content = marker + "  " + body
	}
	return c.cyan("  │") + content + pad(inner-visibleLen(content)) + c.cyan("│")
}

func visibleLen(s string) int {
	n := 0
	for i := 0; i < len(s); {
		if s[i] == '\033' {
			for i < len(s) && s[i] != 'm' {
				i++
			}
			i++
			continue
		}
		_, size := utf8.DecodeRuneInString(s[i:])
		n++
		i += size
	}
	return n
}

func pad(n int) string {
	if n < 0 {
		n = 0
	}
	return fmt.Sprintf("%*s", n, "")
}

type consoleStyle struct {
	color bool
}

func newConsoleStyle(color bool) consoleStyle {
	return consoleStyle{color: color}
}

func (s consoleStyle) wrap(code, text string) string {
	if !s.color || text == "" {
		return text
	}
	return code + text + "\033[0m"
}

func (s consoleStyle) bold(text string) string   { return s.wrap("\033[1m", text) }
func (s consoleStyle) dim(text string) string    { return s.wrap("\033[2m", text) }
func (s consoleStyle) cyan(text string) string   { return s.wrap("\033[36m", text) }
func (s consoleStyle) green(text string) string  { return s.wrap("\033[32m", text) }
func (s consoleStyle) yellow(text string) string { return s.wrap("\033[33m", text) }
func (s consoleStyle) red(text string) string    { return s.wrap("\033[31m", text) }

func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
