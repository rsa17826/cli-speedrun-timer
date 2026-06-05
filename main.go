package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
)

var read *IMan.ManagerConnection
var send *IMan.ManagerConnection

var level = 0
var levelPos = [][]int32{
	{546, 293},
	{549, 555},
	{944, 556},
	{547, 808},
	{935, 814},
}
var started bool
var paused bool
var ended bool
var startTime time.Time
var accumulatedTime time.Duration // Tracks time accumulated across pauses for current split
var elapsed time.Duration         // Tracks current live split time
var totalBest time.Duration
var sob time.Duration
var wrSob time.Duration // Track World Record Sum of Bests

// Split Variables
const totalSplits = 5

// Separate persistent files for distinct category matching
var fullRunPath = "./times_full_run"
var ilPath = "./times_il_all"
var wrPath = "./wrs"

var activeSplit = 0
var splitTimes = make([]time.Duration, totalSplits)

// Distinct tracking arrays to print both contexts concurrently
var fullRunBestTimes = make([]time.Duration, totalSplits)
var ilBestTimes = make([]time.Duration, totalSplits)
var wrTimes = make([]time.Duration, totalSplits)

// IL Mode Config
var ilMode int // 0 means standard full-run, 1-5 indicates specific IL split

// ANSI Color Escape Codes
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Purple = "\033[35m"
	White  = "\033[37m"
)

// Load records across configurations including World Records
func loadBestTimes() {
	// 1. Load Full Run Records
	if file, err := os.Open(fullRunPath); err == nil {
		scanner := bufio.NewScanner(file)
		i := 0
		for i < totalSplits && scanner.Scan() {
			if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
				fullRunBestTimes[i] = time.Duration(ms) * time.Millisecond
			}
			i++
		}
		if scanner.Scan() {
			if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
				totalBest = time.Duration(ms) * time.Millisecond
			}
		}
		file.Close()
	}

	// 2. Load Individual Level (IL) Records
	if file, err := os.Open(ilPath); err == nil {
		scanner := bufio.NewScanner(file)
		i := 0
		sob = 0
		for scanner.Scan() && i < totalSplits {
			if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
				ilBestTimes[i] = time.Duration(ms) * time.Millisecond
				sob += ilBestTimes[i]
			}
			i++
		}
		file.Close()
	}

	// 3. Load World Records (WR)
	if file, err := os.Open(wrPath); err == nil {
		scanner := bufio.NewScanner(file)
		i := 0
		wrSob = 0
		for scanner.Scan() && i < totalSplits {
			if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
				wrTimes[i] = time.Duration(ms) * time.Millisecond
				wrSob += wrTimes[i]
			}
			i++
		}
		file.Close()
	}
}

// Explicitly pass which type of file we are saving to decouple structural rules
func saveFile(target string) {
	switch target {
	case "il":
		file, err := os.Create(ilPath)
		if err != nil {
			return
		}
		defer file.Close()
		for _, t := range ilBestTimes {
			fmt.Fprintf(file, "%d\n", t.Milliseconds())
		}
	case "full":
		file, err := os.Create(fullRunPath)
		if err != nil {
			return
		}
		defer file.Close()
		for _, t := range fullRunBestTimes {
			fmt.Fprintf(file, "%d\n", t.Milliseconds())
		}
		fmt.Fprintf(file, "%d\n", totalBest.Milliseconds())
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func pad(text string, padChar string, totalSize int, bn bool) string {
	if len(text) >= totalSize {
		return text
	}

	totalPadding := totalSize - len(text)
	leftPadding := totalPadding / 2
	rightPadding := totalPadding - leftPadding

	leftStr := strings.Repeat(padChar, leftPadding)
	rightStr := strings.Repeat(padChar, rightPadding)

	if bn {
		return leftStr + text + rightStr + "\n"
	}
	return leftStr + text + rightStr
}

func pt() {
	var size int = 104
	var sb strings.Builder
	sb.WriteString("\033[H") // Return cursor to home

	if ilMode > 0 {
		sb.WriteString(pad(fmt.Sprintf(" INDIVIDUAL LEVEL (IL %d) ", ilMode), "=", size, true))
	} else {
		sb.WriteString(pad(" SPEEDRUN SPLITS ", "=", size, true))
	}

	var currentTotal time.Duration

	for i := range totalSplits {
		if ilMode > 0 && i != ilMode-1 {
			continue
		}

		splitName := fmt.Sprintf("Segment %d:", i+1)
		var dispTime time.Duration
		var segmentColor = Cyan

		targetCompareTime := fullRunBestTimes[i]
		if ilMode > 0 {
			targetCompareTime = ilBestTimes[i]
		}

		if i < activeSplit {
			dispTime = splitTimes[i]
			currentTotal += splitTimes[i]
			if targetCompareTime > 0 {
				if dispTime > targetCompareTime {
					segmentColor = Red
				} else if dispTime > ilBestTimes[i] {
					segmentColor = Yellow
				} else {
					segmentColor = Green
				}
			}
		} else if i == activeSplit && started {
			dispTime = elapsed
			currentTotal += elapsed
			if targetCompareTime > 0 {
				if dispTime > targetCompareTime {
					segmentColor = Red
				} else if dispTime > ilBestTimes[i] {
					segmentColor = Yellow
				} else {
					segmentColor = Green
				}
			} else {
				segmentColor = Cyan
			}
		} else {
			dispTime = 0
		}

		fullBestStr := "NONE"
		if fullRunBestTimes[i] > 0 {
			fullBestStr = formatDuration(fullRunBestTimes[i])
		}

		ilBestStr := "NONE"
		if ilBestTimes[i] > 0 {
			ilBestStr = formatDuration(ilBestTimes[i])
		}

		wrStr := "NONE"
		if wrTimes[i] > 0 {
			wrStr = formatDuration(wrTimes[i])
		}

		timeStr := "--:--:--.---"
		if started && i <= activeSplit || i < activeSplit {
			timeStr = formatDuration(dispTime)
		}

		p := int((float64(ilBestTimes[i]) / float64(fullRunBestTimes[i])) * 100.0)
		pc := Yellow
		if p == 100 {
			pc = Green
		}
		wrp := int((float64(wrTimes[i]) / float64(ilBestTimes[i])) * 100.0)
		wrpc := Yellow
		if wrp == 100 {
			wrpc = Green
		}
		fmt.Fprintf(&sb, "  %-10s Time: %s%-12s%s (Run Best: %s%-12s%s %s%d%%%s IL Best: %s%-12s%s WR: %s%s%s %s%d%%%s)\033[K\n",
			splitName, segmentColor, timeStr, Reset, Purple, fullBestStr, Reset, wrpc, wrp, Reset, Purple, ilBestStr, Reset, Red, wrStr, Reset, pc, p, Reset)
	}

	if ilMode == 0 {
		sb.WriteString(pad("", "-", size, false))
		sb.WriteString("\033[K\n")
		totalColor := Cyan
		if totalBest > 0 && started {
			if currentTotal <= totalBest {
				totalColor = Green
			} else {
				totalColor = Red
			}
		}

		bestTotalStr := "NONE"
		if totalBest > 0 {
			bestTotalStr = formatDuration(totalBest)
		}

		p := int((float64(sob) / float64(totalBest)) * 100.0)
		pc := Yellow
		if p == 100 {
			pc = Green
		}
		wrp := int((float64(wrSob) / float64(sob)) * 100.0)
		wrpc := Yellow
		if wrp == 100 {
			wrpc = Green
		}
		fmt.Fprintf(&sb, "  %-10s Time: %s%-12s%s (Best Total: %s%-12s%s) (SOB: %s%-12s%s %s%d%%%s WR: %s%s%s %s%d%%%s)\033[K\n",
			"TOTAL:", totalColor, formatDuration(currentTotal), Reset, Purple, bestTotalStr, Reset, Purple, formatDuration(sob), Reset, wrpc, wrp, Reset, Red, formatDuration(wrSob), Reset, pc, p, Reset)
	}
	sb.WriteString(pad("", "=", size, false))
	sb.WriteString("\033[K\n")

	// Status Line Footer
	if !started {
		fmt.Fprintf(&sb, "Status: %sSTOPPED / READY%s", White, Reset)
	} else if paused {
		nextText := fmt.Sprintf("Split %d", activeSplit+2)
		if ilMode > 0 {
			nextText = "Done"
		}
		fmt.Fprintf(&sb, "Status: %sPAUSED (Next WASD starts %s)%s", Yellow, nextText, Reset)
	} else {
		fmt.Fprintf(&sb, "Status: %sRUNNING%s", Green, Reset)
	}

	sb.WriteString("\033[J")
	fmt.Print(sb.String())
}

var bi bool
var onmb bool

type WindowTracker struct {
	mu          sync.Mutex
	LastActive  bool
	modifierCmd *exec.Cmd
	pidFile     string
}

func (wt *WindowTracker) listenToHyprland() {
	runtimeDir := os.Getenv("XDG_RUNTIME_DIR")
	instanceSig := os.Getenv("HYPRLAND_INSTANCE_SIGNATURE")
	if runtimeDir == "" || instanceSig == "" {
		fmt.Fprintln(os.Stderr, "Error: Missing Hyprland environment variables.")
		return
	}

	socketPath := fmt.Sprintf("%s/hypr/%s/.socket2.sock", runtimeDir, instanceSig)

	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error connecting to socket: %v\n", err)
		return
	}
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()

		if after, ok := strings.CutPrefix(line, "activewindow>>"); ok {
			payload := after
			parts := strings.Split(payload, ",")
			if len(parts) == 0 {
				continue
			}
			activeClass := parts[0]

			wt.mu.Lock()
			if activeClass == "mathbreakers.exe" {
				if !wt.LastActive {
					wt.handleWindowActive()
					wt.LastActive = true
				}
			} else {
				if wt.LastActive {
					wt.handleWindowInactive()
					wt.LastActive = false
				}
			}
			wt.mu.Unlock()
		}
	}
}

func (wt *WindowTracker) handleWindowActive() {
	exec.Command("hyprctl", "dispatch", "hl.dsp.window.tag({ tag = \"-math_hide\", window = \"class:^Mathbreakers$\" })").Run()

	if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
		wt.modifierCmd.Process.Kill()
	}

	wt.modifierCmd = exec.Command("keyModifier",
		"--modify", "space", "turbo", "downFor", "20ms", "delay", "20ms",
		"--modify", "space", "maxPressTime", "350ms",
		"--modify", "e", "replace", "r",
		"--modify", "2", "replace", "6",
		"--modify", "3", "replace", "6",
		"--modify", "4", "replace", "j",
		"--modify", "4", "turbo",
		"--modify", "rbutton", "turbo", "downFor", "1ms", "delay", "1ms",
		"--modify", "f", "replace", "j",
		"--modify", "rbutton", "replace", "k",
	)

	if err := wt.modifierCmd.Start(); err == nil {
		_ = os.WriteFile(wt.pidFile, fmt.Appendf(nil, "%d", wt.modifierCmd.Process.Pid), 0644)
	}
}

func levelEnded() {
	if started && !paused && !ended {
		paused = true

		finalSegmentTime := accumulatedTime + time.Since(startTime)
		accumulatedTime = finalSegmentTime
		elapsed = finalSegmentTime
		splitTimes[activeSplit] = finalSegmentTime

		if ilBestTimes[activeSplit] == 0 || finalSegmentTime < ilBestTimes[activeSplit] {
			ilBestTimes[activeSplit] = finalSegmentTime
			saveFile("il")
		}

		if ilMode > 0 {
			ended = true
		} else if activeSplit == totalSplits-1 {
			var totalRunTime time.Duration
			for _, t := range splitTimes {
				totalRunTime += t
			}

			if totalBest == 0 || totalRunTime < totalBest {
				totalBest = totalRunTime
				copy(fullRunBestTimes, splitTimes)
				saveFile("full")
			}
			ended = true
		} else {
			go func() {
				time.Sleep(400 * time.Millisecond)
				moveMouse(1920/2, (1080/2)+75)
				click()
				time.Sleep(350 * time.Millisecond)
				moveMouse(596, 223)
				click()
				moveMouse(levelPos[activeSplit+1][0], levelPos[activeSplit+1][1])
				click()
				moveMouse(1920/2, 1080/2)
			}()
		}
	}
}

func (wt *WindowTracker) handleWindowInactive() {
	exec.Command("hyprctl", "dispatch", "hl.dsp.window.tag({ tag = \"+math_hide\", window = \"class:^Mathbreakers$\" })").Run()

	if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
		wt.modifierCmd.Process.Kill()
		wt.modifierCmd = nil
	}
	_ = os.WriteFile(wt.pidFile, []byte("0"), 0644)
}

func (wt *WindowTracker) cleanup() {
	wt.mu.Lock()
	defer wt.mu.Unlock()
	if wt.modifierCmd != nil && wt.modifierCmd.Process != nil {
		wt.modifierCmd.Process.Kill()
	}
	os.Remove(wt.pidFile)
}

func main() {
	var err error
	tracker := &WindowTracker{
		pidFile: "/tmp/mathbreakers_pid_a",
	}
	defer tracker.cleanup()

	go tracker.listenToHyprland()

	fmt.Print("\033[?25l")

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Print("\033[?25h")
		os.Exit(0)
	}()

	flag.IntVar(&ilMode, "il", 0, "Run a single Individual Level (1-5). Standard mode if 0.")
	flag.Parse()

	if ilMode < 0 || ilMode > 5 {
		fmt.Println("Error: -il option must be between 1 and 5")
		os.Exit(1)
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}
	defer watcher.Close()

	watchDir := "/data/games/mathbreakers"
	targetFile := "level_cleared.txt"
	_ = os.Remove(watchDir + "/" + targetFile)
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}

				if event.Has(fsnotify.Create) {
					if filepath.Base(event.Name) == targetFile {
						levelEnded()
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

	err = watcher.Add(watchDir)
	if err != nil {
		log.Fatalf("Failed to add directory to watcher: %v", err)
	}
	loadBestTimes()

	read, err = IMan.Connect(IMan.ModeBlocking)
	if err != nil {
		panic(err)
	}
	send, err = IMan.Connect(IMan.ModeInjection)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(1 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				if !started || paused {
					pt()
					continue
				}

				elapsed = accumulatedTime + time.Since(startTime)
				pt()
			}
		}
	}()

	for {
		ev, err := read.ReadNext()
		if err != nil {
			panic(err)
		}
		if !tracker.LastActive {
			read.BlockInput(0)
			continue
		}
		if ev.Event.Type == input.EV_KEY {
			switch ev.Event.Code {
			case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
				if ev.Event.Value == 1 {
					if ended {
						continue
					}
					if !started {
						started = true
						paused = false

						if ilMode > 0 {
							activeSplit = ilMode - 1
						} else {
							activeSplit = 0
						}

						elapsed = 0
						accumulatedTime = 0
						for i := range splitTimes {
							splitTimes[i] = 0
						}
						startTime = time.Now()
					} else if paused {
						if ilMode == 0 {
							activeSplit++
							elapsed = 0
							accumulatedTime = 0
							paused = false
							startTime = time.Now()
						}
					}
				}
			case input.KEY_TAB:
				if ev.Event.Value == 1 {
					ended = false
					started = false
					paused = false
					activeSplit = 0
					accumulatedTime = 0
					elapsed = 0
				}
			case input.KEY_ESC:
				if ev.Event.Value == 0 {
					ended = false
					started = false
					paused = false
					activeSplit = 0
					accumulatedTime = 0
					elapsed = 0
					for i := range splitTimes {
						splitTimes[i] = 0
					}
					read.BlockInput(0)
					go func() {
						time.Sleep(20 * time.Millisecond)
						bi = true
						moveMouse(1920/2, (1080/2)+75)
						click()
						time.Sleep(350 * time.Millisecond)
						moveMouse(596, 223)
						click()
						if ilMode > 0 {
							moveMouse(levelPos[ilMode-1][0], levelPos[ilMode-1][1])
						} else {
							moveMouse(levelPos[0][0], levelPos[0][1])
						}
						click()
						moveMouse(1920/2, 1080/2)
						bi = false
					}()
					loadBestTimes()
					continue
				}
			}
		}
		if bi {
			read.BlockInput(1)
		} else {
			read.BlockInput(0)
		}
	}
}

func moveMouse(x, y int32) {
	var err error
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_X, Value: -9999})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_Y, Value: -9999})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	time.Sleep(10 * time.Millisecond)
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_X, Value: -10000})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_Y, Value: -10000})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	time.Sleep(10 * time.Millisecond)
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_X, Value: x / 2})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_Y, Value: y / 2})
	if err != nil {
		println(err)
	}
	err = send.Send(IMan.WireEvent{Type: 0, Value: 0, Code: 0})
	if err != nil {
		println(err)
	}
	time.Sleep(10 * time.Millisecond)
}

func playLevel(i int) {
	bi = true
	level = i
	moveMouse(596, 223)
	click()
	moveMouse(levelPos[level][0], levelPos[level][1])
	click()
	moveMouse(1920/2, 1080/2)
	bi = false
}

func click() {
	send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 1})
	send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)

	send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 0})
	send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)
}
