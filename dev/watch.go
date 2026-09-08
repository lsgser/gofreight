package dev

import (
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
)

var watchExtensions = map[string]struct{}{
	".go": {}, ".gft": {}, ".html": {}, ".css": {}, ".sql": {},
	".env": {}, ".yaml": {}, ".yml": {}, ".json": {},
}

var skipDirNames = map[string]struct{}{
	".git": {}, "node_modules": {}, "vendor": {}, "tmp": {},
	"storage": {}, "dist": {}, "build": {},
}

// Watch reruns a command when source or template files change (nodemon-style dev reload).
func Watch(root, command string, args ...string) error {
	root, err := filepath.Abs(root)
	if err != nil {
		return err
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}
	defer watcher.Close()

	if err := watchTree(watcher, root); err != nil {
		return err
	}

	log.Printf("Hot reload enabled — watching %s (Ctrl+C to stop)", root)

	var (
		mu      sync.Mutex
		current *exec.Cmd
	)

	start := func() {
		mu.Lock()
		defer mu.Unlock()

		if current != nil && current.Process != nil {
			_ = current.Process.Signal(syscall.SIGTERM)
			done := make(chan struct{})
			go func(cmd *exec.Cmd) {
				_ = cmd.Wait()
				close(done)
			}(current)
			select {
			case <-done:
			case <-time.After(3 * time.Second):
				_ = current.Process.Kill()
			}
		}

		cmd := exec.Command(command, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Dir = root
		if err := cmd.Start(); err != nil {
			log.Printf("failed to start %s: %v", command, err)
			return
		}
		current = cmd
		log.Printf("started %s %s (pid %d)", command, strings.Join(args, " "), cmd.Process.Pid)
	}

	start()

	debounce := time.NewTimer(0)
	if !debounce.Stop() {
		<-debounce.C
	}

	restart := func() {
		if !debounce.Stop() {
			select {
			case <-debounce.C:
			default:
			}
		}
		debounce.Reset(300 * time.Millisecond)
	}

	go func() {
		for range debounce.C {
			log.Printf("change detected — restarting...")
			start()
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sig)

	for {
		select {
		case err := <-watcher.Errors:
			log.Printf("watch error: %v", err)
		case event := <-watcher.Events:
			if !shouldRestart(event.Name) {
				continue
			}
			if event.Has(fsnotify.Chmod) && !event.Has(fsnotify.Write|fsnotify.Create|fsnotify.Remove|fsnotify.Rename) {
				continue
			}
			restart()
		case <-sig:
			mu.Lock()
			if current != nil && current.Process != nil {
				_ = current.Process.Signal(syscall.SIGTERM)
				_ = current.Wait()
			}
			mu.Unlock()
			return nil
		}
	}
}

func watchTree(watcher *fsnotify.Watcher, root string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if path != root && shouldSkipDir(d.Name()) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	})
}

func shouldSkipDir(name string) bool {
	_, ok := skipDirNames[name]
	return ok
}

func shouldRestart(path string) bool {
	ext := filepath.Ext(path)
	if _, ok := watchExtensions[ext]; ok {
		return true
	}
	base := filepath.Base(path)
	return base == ".env" || strings.HasPrefix(base, ".env.")
}
