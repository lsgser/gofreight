package dev

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

// Watch reruns a command when Go or template files change (development helper).
func Watch(root, command string, args ...string) error {
	log.Printf("Watching %s — press Ctrl+C to stop", root)
	var lastRun time.Time

	for {
		changed := false
		filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
			if err != nil || info.IsDir() {
				return nil
			}
			ext := filepath.Ext(path)
			if ext != ".go" && ext != ".html" && ext != ".sql" && ext != ".css" {
				return nil
			}
			if info.ModTime().After(lastRun) && time.Since(info.ModTime()) < 2*time.Second {
				changed = true
			}
			return nil
		})

		if changed || lastRun.IsZero() {
			lastRun = time.Now()
			cmd := exec.Command(command, args...)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			cmd.Run()
		}
		time.Sleep(500 * time.Millisecond)
	}
}
