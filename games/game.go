// Package games holds one file per supported game. Each file registers a
// *Game in init(); the engine in package main looks the game up by name.
package games

import (
	"fmt"
	"sort"
	"time"

	input "github.com/rsa17826/go-input-lib"
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
	Block   bool         // true = swallow the original key press
	OnPress func(c Ctx) // runs in its own goroutine, so sleeping is fine
}

// Game describes everything the engine needs to time one game.
// Every field is required except Binds (which may be empty).
type Game struct {
	Name        string   // name used on the command line and for the input-manager connection
	WindowClass string   // Hyprland window class of the game, e.g. "mathbreakers.exe"
	Splits      []string // split names, in order; len(Splits) is the number of splits

	DataDir string // directory holding times_full_run / times_il_all / wrs

	// File-based signalling from the game/mod.
	WatchDir       string // directory the game/mod writes into
	LevelDoneFile  string // created by the game when a split is finished
	LevelResetFile string // created by the game when the level was reset

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
	case g.WindowClass == "":
		panic(fmt.Sprintf("games: %s: WindowClass is required", g.Name))
	case len(g.Splits) == 0:
		panic(fmt.Sprintf("games: %s: Splits is required", g.Name))
	case g.DataDir == "":
		panic(fmt.Sprintf("games: %s: DataDir is required", g.Name))
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

// Re-exported so game files can reach key codes without a second import
// if they prefer; they can also import go-input-lib directly.
var _ = input.EV_KEY
