package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

// setupWatcher creates and starts a file-system watcher on the game's WatchDir.
// When the game drops LevelDoneFile, levelEnded is called and the file is removed.
// The caller is responsible for closing the returned watcher.
func (a *App) setupWatcher() (*fsnotify.Watcher, error) {
	g := a.rs.Game

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	// Remove any stale trigger file from a previous session.
	_ = os.Remove(filepath.Join(g.WatchDir, g.LevelDoneFile))

	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Has(fsnotify.Create) {
					if filepath.Base(event.Name) == g.LevelDoneFile {
						a.levelEnded()
						time.Sleep(50 * time.Millisecond)
						os.Remove(event.Name)
					}
					if filepath.Base(event.Name) == g.LevelResetFile {
						rs := a.rs
						if !(rs.ILMode == 0 && rs.NoReset) {
							rs.Reset()
						}

						loadBestTimes(rs)
						time.Sleep(50 * time.Millisecond)
						os.Remove(event.Name)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Printf("Watcher error: %v", err)
			}
		}
	}()

	if err := watcher.Add(g.WatchDir); err != nil {
		watcher.Close()
		return nil, err
	}

	return watcher, nil
}
