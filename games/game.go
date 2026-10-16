// Package games holds one file per supported game. Each file registers a
// *Game in init(); the engine in package main looks the game up by name.
package games

import (
	"fmt"
	"path/filepath"
	"runtime"
	"sort"
	"time"

	"github.com/rsa17826/input-manager/IMan"
)

// Ctx is what a game's macros can do. The engine's App implements it.
type Ctx interface {
	// Send injects a raw input event (call Sync afterwards to flush).
	Send(ev IMan.WireEvent) error
	// Sync flushes pending injected events.
	Sync()
	// MoveMouse moves the cursor to absolute coordinates inside the game monitor.
	MoveMouse(x, y int32)
	// Click presses and releases the left mouse button.
	Click()
	// TapKey presses a key, holds it for `hold`, then releases it.
	TapKey(code uint16, hold time.Duration)
	// Scroll sends a mouse wheel event.
	Scroll(amount int32)
	// ActiveSplit returns the 0-indexed split currently being timed.
	ActiveSplit() int
}

// Bind is a custom key macro. It fires on key press while the game is focused.
type Bind struct {
	Key     uint16
	Block   bool        // true = swallow the original key press
	OnPress func(c Ctx) // runs in its own goroutine, so sleeping is fine
}

// Game describes everything the engine needs to time one game.
// Every field is required except Binds (which may be empty) and whichever of
// WindowClass / WindowTitle is not used.
type Game struct {
	Name        string   // name used on the command line and for the input-manager connection
	WindowClass string   // Hyprland window class of the game, e.g. "mathbreakers.exe" (set this OR WindowTitle)
	WindowTitle string   // Hyprland window title of the game (set this OR WindowClass)
	Splits      []string // split names, in order; len(Splits) is the number of splits

	// File-based signalling from the game/mod.
	WatchDir       string // directory the game/mod writes into
	LevelDoneFile  string // created by the game when a split is finished
	LevelResetFile string // created by the game when the level was reset

	// NoPause: when true the timer runs straight into the next split when one
	// ends, instead of pausing until a start key is pressed.
	NoPause bool

	StartKeys   []uint16 // first press starts the timer / resumes the next split
	ResetKeys   []uint16 // reset the run (key still passes through)
	RestartKeys []uint16 // blocked; reset (unless -nor) and call PlayLevel
	Binds       []Bind   // extra macros

	// PlayLevel navigates to / launches the 0-indexed split.
	PlayLevel func(c Ctx, split int)
}

var registry = map[string]*Game{}

// Register adds a game. It panics on missing required data so mistakes in a
// game file show up immediately at startup.
func Register(g *Game) {
	switch {
	case g.Name == "":
		panic("games: Game.Name is required")
	case (g.WindowClass == "") == (g.WindowTitle == ""):
		panic(fmt.Sprintf("games: %s: set exactly one of WindowClass or WindowTitle", g.Name))
	case len(g.Splits) == 0:
		panic(fmt.Sprintf("games: %s: Splits is required", g.Name))
	case g.WatchDir == "":
		panic(fmt.Sprintf("games: %s: WatchDir is required", g.Name))
	case g.LevelDoneFile == "":
		panic(fmt.Sprintf("games: %s: LevelDoneFile is required", g.Name))
	case g.LevelResetFile == "":
		panic(fmt.Sprintf("games: %s: LevelResetFile is required", g.Name))
	case len(g.StartKeys) == 0:
		panic(fmt.Sprintf("games: %s: StartKeys is required", g.Name))
	case g.PlayLevel == nil:
		panic(fmt.Sprintf("games: %s: PlayLevel is required", g.Name))
	}
	if _, dup := registry[g.Name]; dup {
		panic(fmt.Sprintf("games: %s registered twice", g.Name))
	}
	registry[g.Name] = g
}

// WindowMatch returns which Hyprland window property identifies the game
// ("class" or "title") and the value it must equal.
func (g *Game) WindowMatch() (field, value string) {
	if g.WindowClass != "" {
		return "class", g.WindowClass
	}
	return "title", g.WindowTitle
}

// Get returns the game registered under name.
func Get(name string) (*Game, bool) {
	g, ok := registry[name]
	return g, ok
}

// Names lists all registered games, sorted.
func Names() []string {
	names := make([]string, 0, len(registry))
	for n := range registry {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

// Dir returns the directory of the Go file that calls it. A game calls this
// from its own package so its times and other files can live next to its code,
// no matter which directory the timer is run from.
func Dir() string {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("games.Dir: could not determine caller file")
	}
	return filepath.Dir(file)
}
