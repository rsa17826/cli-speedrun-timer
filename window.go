package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"

	"speedrun/games"
)

// WindowTracker monitors the active Hyprland window and records whether the
// game's window is focused.
type WindowTracker struct {
	mu         sync.Mutex
	LastActive bool
	field      string // "class" or "title"
	value      string
}

// NewWindowTracker returns a tracker for the game's window.
func NewWindowTracker(g *games.Game) *WindowTracker {
	field, value := g.WindowMatch()
	return &WindowTracker{field: field, value: value}
}

// listenToHyprland connects to the Hyprland IPC socket and tracks focus changes.
func (wt *WindowTracker) listenToHyprland() {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	instanceSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if runtimeDir == "" || instanceSig == "" {
		fmt.Fprintln(os.Stderr, "Error: missing Hyprland environment variables (XDG_RUNTIME_DIR, HYPRLAND_INSTANCE_SIGNATURE)")
		return
	}

	socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", runtimeDir, instanceSig)
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to Hyprland socket: %v\n", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		after, ok := strings.CutPrefix(scanner.Text(), "activewindow>>")
		if !ok {
			continue
		}

		// Format is "class,title"; the title itself may contain commas.
		parts := strings.SplitN(after, ",", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintf("unexpected activewindow event: %q", after))
		}
		class, title := parts[0], parts[1]

		active := title
		if wt.field == "class" {
			active = class
		}

		wt.mu.Lock()
		wt.LastActive = active == wt.value
		wt.mu.Unlock()
	}
}
