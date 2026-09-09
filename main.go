package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
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
	rs := NewRunState()

	flag.IntVar(&rs.ILMode, "il", 0, "Individual Level mode: run a single split (1-5). 0 = full run.")
	flag.BoolVar(&rs.NoReset, "nor", false, "No-reset mode: ESC navigates back without resetting the run.")
	flag.Parse()

	if rs.ILMode < 0 || rs.ILMode > TotalSplits {
		fmt.Fprintf(os.Stderr, "Error: -il must be between 0 and %d\n", TotalSplits)
		os.Exit(1)
	}

	app := &App{
		rs:      rs,
		tracker: NewWindowTracker(),
	}

	var err error
	app.read, err = IMan.Connect("mathbreakers", IMan.ModeFilter)
	if err != nil {
		log.Fatalf("Failed to connect input reader: %v", err)
	}
	app.send, err = IMan.Connect("mathbreakers", IMan.ModeInjection)
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

		if ev.Event.Type == input.EV_KEY {
			switch ev.Event.Code {
			case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
				if ev.Event.Value == 1 {
					a.handleMovementKey()
				}
			case input.KEY_TAB:
				if ev.Event.Value == 1 {
					a.rs.Reset()
				}
			case input.KEY_ESC:
				if ev.Event.Value == 0 {
					a.read.BlockInput(ev.Event.Seq, 0)
					go a.handleEscape()
					continue
				}
			case input.BTN_RIGHT:
				if ev.Event.Value == 1 {
					print(a.rs.ActiveSplit)
					if a.rs.ActiveSplit == 4 {
						a.read.BlockInput(ev.Event.Seq, 1)
						a.send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_WHEEL, Value: 1})
					} else {
						a.read.BlockInput(ev.Event.Seq, 1)
						go func() {
							a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.KEY_R, Value: 1})
							time.Sleep(40 * time.Millisecond)
							a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.KEY_R, Value: 0})
						}()
					}
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

// handleMovementKey starts or continues a run when WASD is pressed.
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
	if !(rs.ILMode == 0 && rs.NoReset) {
		rs.Reset()
	}

	go func() {
		time.Sleep(20 * time.Millisecond)
		a.blocking = true
		a.clickExitLevelButton()
		time.Sleep(600 * time.Millisecond)
		a.moveMouse(735, 963)
		a.click()
		time.Sleep(400 * time.Millisecond)

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

		// a.blocking = true
		// time.Sleep(800 * time.Millisecond)

		// for range 11 {
		// 	a.sendRel(input.REL_X, 1)
		// 	a.sendSync()
		// 	time.Sleep(10 * time.Millisecond)
		// }
		// a.blocking = false
	}()

	loadBestTimes(rs)
}

// levelEnded is called when the game signals a level completion.
func (a *App) levelEnded() {
	rs := a.rs
	if !rs.Started || rs.Paused || rs.Ended {
		return
	}

	rs.Paused = true
	finalTime := rs.AccumulatedTime + time.Since(rs.StartTime)
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

	if rs.ActiveSplit == TotalSplits-1 {
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
		// Auto-advance to the next level after a short delay.
		go func() {
			time.Sleep(400 * time.Millisecond)
			a.clickExitLevelButton()
			time.Sleep(600 * time.Millisecond)
			a.playLevel(rs.ActiveSplit + 1)
			a.moveMouse(ScreenWidth/2, ScreenHeight/2)
			a.moveMouse(ScreenWidth/2, ScreenHeight/2)
			a.moveMouse(ScreenWidth/2, ScreenHeight/2)
			a.moveMouse(ScreenWidth/2, ScreenHeight/2)
		}()
	}
}
