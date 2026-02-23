package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

var (
	currentCmd *exec.Cmd
	cmdLock    sync.Mutex
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	// Start initial process
	go func() {
		restartProcess()
	}()

	// Watch current directory recursively
	if err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if shouldIgnore(path) {
				return filepath.SkipDir
			}
			return watcher.Add(path)
		}
		return nil
	}); err != nil {
		log.Fatal(err)
	}

	done := make(chan bool)

	// Debounce timer
	var timer *time.Timer
	var timerLock sync.Mutex

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if shouldIgnore(event.Name) {
					continue
				}
				if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create || event.Op&fsnotify.Remove == fsnotify.Remove || event.Op&fsnotify.Rename == fsnotify.Rename {
					log.Println("Modified file:", event.Name)

					timerLock.Lock()
					if timer != nil {
						timer.Stop()
					}
					timer = time.AfterFunc(500*time.Millisecond, func() {
						restartProcess()
					})
					timerLock.Unlock()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("error:", err)
			}
		}
	}()

	log.Println("Watching for file changes...")
	<-done
}

func shouldIgnore(path string) bool {
	// Ignore hidden files/dirs
	if strings.Contains(path, "/.") || strings.HasPrefix(path, ".") {
		return path != "."
	}
	// Ignore build artifacts and vendor
	if strings.Contains(path, "vendor") || strings.Contains(path, "bin") || strings.Contains(path, "tmp") || strings.Contains(path, "ghost-ops") || strings.Contains(path, "ghost-dev") {
		// ghost-ops is the binary name, ghost-dev is this tool
		// If path contains ghost-ops, it might be the binary or source dir?
		// "cmd/ghost-ops" is source. "ghost-ops" is binary.
		// Better check strict equality for binary names or extensions.
		if path == "ghost-ops" || path == "ghost-dev" {
			return true
		}
	}

	return false
}

func restartProcess() {
	cmdLock.Lock()
	defer cmdLock.Unlock()

	if currentCmd != nil && currentCmd.Process != nil {
		log.Println("Killing old process...")
		if err := currentCmd.Process.Kill(); err != nil {
			log.Println("Failed to kill process:", err)
		}
		// Wait ensures resources are released, but Process.Kill is async signal.
		// We don't want to block indefinitely.
		currentCmd.Wait()
	}

	log.Println("Building...")
	buildCmd := exec.Command("go", "build", "-o", "ghost-ops", "cmd/ghost-ops/main.go")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		log.Println("Build failed:", err)
		return
	}

	log.Println("Starting process...")
	cmd := exec.Command("./ghost-ops")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Println("Failed to start process:", err)
		return
	}
	currentCmd = cmd
}
