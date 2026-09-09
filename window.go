package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
)

// WindowTracker monitors the active Hyprland window and manages the input
// modifier process that applies per-game key remappings.
type WindowTracker struct {
	mu          sync.Mutex
	LastActive  bool
	modifierCmd *exec.Cmd
	pidFile     string
}

// NewWindowTracker returns a tracker configured with the standard PID file path.
func NewWindowTracker() *WindowTracker {
	return &WindowTracker{pidFile: PIDFile}
}

// listenToHyprland connects to the Hyprland IPC socket and reacts to
// window-focus changes, enabling/disabling game-specific key modifiers.
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
		line := scanner.Text()

		after, ok := strings.CutPrefix(line, "activewindow>>")
		if !ok {
			continue
		}

		parts := strings.SplitN(after, ",", 2)
		if len(parts) == 0 {
			continue
		}
		activeClass := parts[0]

		wt.mu.Lock()
		isGame := activeClass == "explorer.exe"
		if isGame && !wt.LastActive {
			wt.LastActive = true
		} else if !isGame && wt.LastActive {
			wt.LastActive = false
		}
		wt.mu.Unlock()
	}
}
