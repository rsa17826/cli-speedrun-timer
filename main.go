package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"

	"speedrun/games"
	"speedrun/winrules"
)

// App ties together all runtime components of the speedrun timer.
type App struct {
	rs       *RunState
	tracker  *WindowTracker
	read     *IMan.ManagerConnection
	send     *IMan.ManagerConnection
	blocking bool // when true, raw inputs are suppressed
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: go run . <game> [-il N] [-nor]\nAvailable games: %s\n", strings.Join(games.Names(), ", "))
		os.Exit(1)
	}
	game, ok := games.Get(os.Args[1])
	if !ok {
		fmt.Fprintf(os.Stderr, "Unknown game %q. Available games: %s\n", os.Args[1], strings.Join(games.Names(), ", "))
		os.Exit(1)
	}

	rs := NewRunState(game)

	fs := flag.NewFlagSet(game.Name, flag.ExitOnError)
	fs.IntVar(&rs.ILMode, "il", 0, "Individual Level mode: run a single split (1-N). 0 = full run.")
	fs.BoolVar(&rs.NoReset, "nor", false, "No-reset mode: ESC navigates back without resetting the run.")
	fs.BoolVar(&rs.NoPause, "np", false, "No-pause mode: after a split ends the timer continues straight into the next split.")
	fs.Parse(os.Args[2:])

	if rs.ILMode < 0 || rs.ILMode > rs.Total {
		fmt.Fprintf(os.Stderr, "Error: -il must be between 0 and %d\n", rs.Total)
		os.Exit(1)
	}

	app := &App{
		rs:      rs,
		tracker: NewWindowTracker(game),
	}

	field, value := game.WindowMatch()
	rulesPath := filepath.Join(game.DataDir, "windowrules.lua")
	rules := winrules.Generate(game.Name, field, value, displayLines(rs.Total)*winrules.LineHeight)
	if err := os.WriteFile(rulesPath, []byte(rules), 0644); err != nil {
		log.Fatalf("Failed to write %s: %v", rulesPath, err)
	}

	var err error
	app.read, err = IMan.Connect(game.Name, IMan.ModeFilter)
	if err != nil {
		log.Fatalf("Failed to connect input reader: %v", err)
	}
	app.send, err = IMan.Connect(game.Name, IMan.ModeInjection)
	if err != nil {
		log.Fatalf("Failed to connect input sender: %v", err)
	}

	// Hide terminal cursor for a clean display.
	fmt.Print("\033[?25l")
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Print("\033[?25h")
		app.read.Close()
		app.send.Close()
		os.Exit(0)
	}()

	go app.tracker.listenToHyprland()

	watcher, err := app.setupWatcher()
	if err != nil {
		log.Fatalf("Failed to set up file watcher: %v", err)
	}
	defer watcher.Close()

	loadBestTimes(rs)

	// Display ticker: updates elapsed time and redraws every millisecond.
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	go func() {
		for range ticker.C {
			if rs.Started && !rs.Paused {
				rs.Elapsed = rs.AccumulatedTime + time.Since(rs.StartTime)
			}
			Render(rs)
		}
	}()

	app.eventLoop()
}

// eventLoop is the main input-processing loop.
func (a *App) eventLoop() {
	g := a.rs.Game
	for {
		ev, err := a.read.ReadNext()
		if err != nil {
			log.Fatalf("Input read error: %v", err)
		}

		// If the game window is not focused, pass all inputs through.
		if !a.tracker.LastActive {
			a.read.BlockInput(ev.Event.Seq, 0)
			continue
		}

		if ev.Event.Type == input.EV_KEY && ev.Event.Value == 1 {
			code := ev.Event.Code
			switch {
			case slices.Contains(g.StartKeys, code):
				a.handleMovementKey()
			case slices.Contains(g.ResetKeys, code):
				a.rs.Reset()
			case slices.Contains(g.RestartKeys, code):
				a.read.BlockInput(ev.Event.Seq, 1)
				a.handleEscape()
				continue
			default:
				handled := false
				for _, b := range g.Binds {
					if b.Key != code {
						continue
					}
					handled = true
					if b.Block {
						a.read.BlockInput(ev.Event.Seq, 1)
						go b.OnPress(a)
					} else {
						a.read.BlockInput(ev.Event.Seq, 0)
						go b.OnPress(a)
					}
					break
				}
				if handled {
					continue
				}
			}
		}

		if a.blocking {
			a.read.BlockInput(ev.Event.Seq, 1)
		} else {
			a.read.BlockInput(ev.Event.Seq, 0)
		}
	}
}

// handleMovementKey starts or continues a run when a start key is pressed.
func (a *App) handleMovementKey() {
	rs := a.rs
	if rs.Ended {
		return
	}
	if !rs.Started {
		rs.Started = true
		rs.Paused = false
		if rs.ILMode > 0 {
			rs.ActiveSplit = rs.ILMode - 1
		} else {
			rs.ActiveSplit = 0
		}
		rs.Elapsed = 0
		rs.AccumulatedTime = 0
		for i := range rs.SplitTimes {
			rs.SplitTimes[i] = 0
		}
		rs.StartTime = time.Now()
	} else if rs.Paused && rs.ILMode == 0 {
		// Advance to the next split.
		rs.ActiveSplit++
		rs.Elapsed = 0
		rs.AccumulatedTime = 0
		rs.Paused = false
		rs.StartTime = time.Now()
	}
}

// handleEscape navigates back to the level select and optionally resets the run.
func (a *App) handleEscape() {
	rs := a.rs
	if !(rs.ILMode == 0 && (rs.NoReset && rs.Level != 0)) {
		rs.Reset()
	}

	switch {
	case rs.ILMode > 0:
		a.playLevel(rs.ILMode - 1) // playLevel is 0-indexed
	case rs.NoReset:
		if rs.Ended {
			a.playLevel(0)
		} else {
			a.playLevel(rs.Level)
		}
	default:
		a.playLevel(0)
	}

	loadBestTimes(rs)
}

// levelEnded is called when the game signals a level completion.
func (a *App) levelEnded() {
	rs := a.rs
	if !rs.Started || rs.Paused || rs.Ended {
		return
	}

	now := time.Now()
	rs.Paused = true
	finalTime := rs.AccumulatedTime + now.Sub(rs.StartTime)
	rs.AccumulatedTime = finalTime
	rs.Elapsed = finalTime
	rs.SplitTimes[rs.ActiveSplit] = finalTime

	// Always update the IL best for this split.
	if rs.ILBestTimes[rs.ActiveSplit] == 0 || finalTime < rs.ILBestTimes[rs.ActiveSplit] {
		rs.ILBestTimes[rs.ActiveSplit] = finalTime
		saveFile(rs, "il")
	}

	if rs.ILMode > 0 {
		rs.Ended = true
		return
	}

	if rs.ActiveSplit == rs.Total-1 {
		// Full run complete — check for a new personal best.
		var total time.Duration
		for _, t := range rs.SplitTimes {
			total += t
		}
		if rs.TotalBest == 0 || total < rs.TotalBest {
			rs.TotalBest = total
			copy(rs.FullRunBestTimes, rs.SplitTimes)
			saveFile(rs, "full")
		}
		rs.Ended = true
	} else {
		next := rs.ActiveSplit + 1
		if rs.NoPause {
			// Keep timing: the next split starts at the instant this one ended.
			rs.AccumulatedTime = 0
			rs.Elapsed = 0
			rs.StartTime = now
			rs.ActiveSplit = next
			rs.Paused = false
		}
		// Auto-advance to the next level after a short delay.
		go func() {
			time.Sleep(200 * time.Millisecond)
			a.playLevel(next)
		}()
	}
}
