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
			wt.handleWindowActive()
			wt.LastActive = true
		} else if !isGame && wt.LastActive {
			wt.handleWindowInactive()
			wt.LastActive = false
		}
		wt.mu.Unlock()
	}
}

// handleWindowActive shows the game window and starts the key modifier process.
func (wt *WindowTracker) handleWindowActive() {
	exec.Command("hyprctl", "dispatch",
		`hl.dsp.window.tag({ tag = "-HIDE", window = "class:^TIMER$" })`).Run()

	// Kill any existing modifier before starting a fresh one.
	// if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
	// 	wt.modifierCmd.Process.Kill()
	// }

	// wt.modifierCmd = exec.Command("keyModifier",
	// 	"--modify", "space", "turbo", "downFor", "20ms", "delay", "20ms",
	// 	"--modify", "space", "maxPressTime", "600ms",
	// 	"--modify", "e", "replace", "r",
	// 	"--modify", "2", "replace", "6",
	// 	"--modify", "3", "replace", "6",
	// 	"--modify", "4", "replace", "j",
	// 	"--modify", "4", "turbo",
	// 	"--modify", "e", "turbo", "downFor", "5ms", "delay", "5ms",
	// 	"--modify", "e", "maxPressTime", "20ms",
	// 	"--modify", "r", "turbo", "downFor", "5ms", "delay", "5ms",
	// 	"--modify", "r", "maxPressTime", "20ms",
	// 	"--modify", "rbutton", "turbo", "downFor", "5ms", "delay", "5ms",
	// 	"--modify", "rbutton", "maxPressTime", "20ms",
	// 	"--modify", "f", "replace", "j",
	// 	"--modify", "rbutton", "replace", "r",
	// )

	// if err := wt.modifierCmd.Start(); err == nil {
	// 	_ = os.WriteFile(wt.pidFile, fmt.Appendf(nil, "%d", wt.modifierCmd.Process.Pid), 0644)
	// }
}

// handleWindowInactive hides the game window and kills the key modifier.
func (wt *WindowTracker) handleWindowInactive() {
	exec.Command("hyprctl", "dispatch",
		`hl.dsp.window.tag({ tag = "+HIDE", window = "class:^TIMER$" })`).Run()

	// if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
	// 	wt.modifierCmd.Process.Kill()
	// 	wt.modifierCmd = nil
	// }
	// _ = os.WriteFile(wt.pidFile, []byte("0"), 0644)
}

// cleanup kills any running modifier and removes the PID file.
// Call via defer in main.
func (wt *WindowTracker) cleanup() {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	// if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
	// 	wt.modifierCmd.Process.Kill()
	// }
	// os.Remove(wt.pidFile)
}
