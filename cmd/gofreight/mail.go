package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/lsgser/gofreight/mail"
)

func registerMailCommands() {
	register(command{
		Name:        "mail:preview",
		Category:    "app",
		Description: "Render a mailable email template to stdout or a file",
		Usage:       "gofreight mail:preview <name> [flags]",
		Run:         handleMailPreview,
	})
}

func handleMailPreview(args []string) {
	opts, name, help := parseMailPreviewArgs(args)
	if help {
		printMailPreviewHelp()
		return
	}
	if name == "" {
		fail(fmt.Errorf("mail:preview requires a mailable or view name (e.g. WelcomeEmail or mail/welcome_email)"))
	}

	viewsRoot := opts.ViewsRoot
	if viewsRoot == "" {
		viewsRoot = mail.DefaultViewsRoot
	}
	if _, err := os.Stat(viewsRoot); err != nil {
		fail(fmt.Errorf("views directory not found: %s (run from your app root)", viewsRoot))
	}

	templateName := mail.ResolvePreviewName(name)
	html, err := mail.RenderView(viewsRoot, templateName, opts.Data)
	if err != nil {
		fail(err)
	}

	if opts.OutFile != "" {
		if err := os.WriteFile(opts.OutFile, []byte(html), 0644); err != nil {
			fail(err)
		}
		fmt.Printf("Wrote %s (%d bytes)\n", opts.OutFile, len(html))
		if opts.Open {
			openFile(opts.OutFile)
		}
		return
	}

	if opts.Open {
		tmp, err := os.CreateTemp("", "gofreight-mail-*.html")
		if err != nil {
			fail(err)
		}
		path := tmp.Name()
		if _, err := tmp.WriteString(html); err != nil {
			tmp.Close()
			os.Remove(path)
			fail(err)
		}
		tmp.Close()
		fmt.Printf("Preview written to %s\n", path)
		openFile(path)
		return
	}

	fmt.Print(html)
}

type mailPreviewOpts struct {
	ViewsRoot string
	OutFile   string
	Open      bool
	Data      map[string]any
}

func parseMailPreviewArgs(args []string) (mailPreviewOpts, string, bool) {
	var opts mailPreviewOpts
	var name string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "-h", "--help", "help":
			return opts, "", true
		case "--views":
			if i+1 >= len(args) {
				fail(fmt.Errorf("--views requires a path"))
			}
			i++
			opts.ViewsRoot = args[i]
		case "--out", "-o":
			if i+1 >= len(args) {
				fail(fmt.Errorf("--out requires a file path"))
			}
			i++
			opts.OutFile = args[i]
		case "--open":
			opts.Open = true
		case "--data":
			if i+1 >= len(args) {
				fail(fmt.Errorf("--data requires JSON"))
			}
			i++
			if err := json.Unmarshal([]byte(args[i]), &opts.Data); err != nil {
				fail(fmt.Errorf("invalid --data JSON: %w", err))
			}
		default:
			if strings.HasPrefix(arg, "-") {
				fail(fmt.Errorf("unknown flag: %s", arg))
			}
			if name == "" {
				name = arg
			}
		}
	}
	return opts, name, false
}
func openFile(path string) {
	abs, _ := filepath.Abs(path)
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", abs)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", abs)
	default:
		cmd = exec.Command("xdg-open", abs)
	}
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Could not open browser: %v\n", err)
	}
}

func printMailPreviewHelp() {
	fmt.Println(`Render a mailable email template without sending mail.

Usage:
  gofreight mail:preview <name> [flags]

Arguments:
  name    Mailable class (WelcomeEmail), snake name (welcome_email), or view path (mail/welcome_email)

Flags:
      --views PATH   Views root (default: app/views)
      --data JSON    Template data as JSON (default: {})
      --out, -o FILE Write HTML to a file instead of stdout
      --open         Write to a temp file and open in the default browser

Examples:
  gofreight mail:preview WelcomeEmail
  gofreight mail:preview mail/welcome_email --data '{"Name":"Ada"}'
  gofreight mail:preview WelcomeEmail --out /tmp/welcome.html --open

Generate mailables: gofreight make:mail WelcomeEmail`)
}
