package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
	"sync"
)

// WindowTracker monitors the active Hyprland window and records whether the
// game's window is focused.
type WindowTracker struct {
	mu          sync.Mutex
	LastActive  bool
	windowClass string
}

// NewWindowTracker returns a tracker for the given Hyprland window class.
func NewWindowTracker(windowClass string) *WindowTracker {
	return &WindowTracker{windowClass: windowClass}
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

		activeClass := strings.SplitN(after, ",", 2)[0]

		wt.mu.Lock()
		wt.LastActive = activeClass == wt.windowClass
		wt.mu.Unlock()
	}
}
