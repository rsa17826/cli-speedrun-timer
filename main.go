package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

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

// Split Variables
const totalSplits = 5

var filePath = "./times"

var activeSplit = 0
var splitTimes = make([]time.Duration, totalSplits)
var bestTimes = make([]time.Duration, totalSplits)

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

// Load best times from file
func loadBestTimes() {
	file, err := os.Open(filePath)
	if err != nil {
		return // File doesn't exist yet, ignore
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if ilMode > 0 {
		if scanner.Scan() {
			ms, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err == nil {
				bestTimes[ilMode-1] = time.Duration(ms) * time.Millisecond
			}
		}
	} else {
		i := 0
		for scanner.Scan() && i < totalSplits {
			ms, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err == nil {
				bestTimes[i] = time.Duration(ms) * time.Millisecond
			}
			i++
		}
	}
}

// Save best times to file
func saveBestTimes() {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()
	if ilMode > 0 {
		fmt.Fprintf(file, "%d\n", bestTimes[ilMode-1].Milliseconds())
	} else {
		for _, t := range bestTimes {
			fmt.Fprintf(file, "%d\n", t.Milliseconds())
		}
	}
}

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func pt() {
	// Use a StringBuilder to buffer the output into a single write operation.
	// This prevents the terminal from rendering line-by-line.
	var sb strings.Builder

	// Move cursor to top-left corner (0,0) WITHOUT clearing the screen
	sb.WriteString("\033[H")

	if ilMode > 0 {
		sb.WriteString(fmt.Sprintf("============== INDIVIDUAL LEVEL (IL %d) ==============\n", ilMode))
	} else {
		sb.WriteString("================== SPEEDRUN SPLITS ==================\n")
	}

	var currentTotal time.Duration
	var bestTotal time.Duration

	for i := range totalSplits {
		if ilMode > 0 && i != ilMode-1 {
			continue
		}

		splitName := fmt.Sprintf("Segment %d:", i+1)
		bestTotal += bestTimes[i]

		var dispTime time.Duration
		var segmentColor = Cyan

		if i < activeSplit {
			dispTime = splitTimes[i]
			currentTotal += splitTimes[i]
			if bestTimes[i] > 0 {
				if dispTime <= bestTimes[i] {
					segmentColor = Green
				} else {
					segmentColor = Red
				}
			}
		} else if i == activeSplit && started {
			dispTime = elapsed
			currentTotal += elapsed
			if bestTimes[i] > 0 {
				if dispTime <= bestTimes[i] {
					segmentColor = Green
				} else {
					segmentColor = Red
				}
			} else {
				segmentColor = Yellow
			}
		} else {
			dispTime = 0
		}

		bestStr := "NONE"
		if bestTimes[i] > 0 {
			bestStr = formatDuration(bestTimes[i])
		}

		timeStr := "--:--:--.---"
		if started && i <= activeSplit || i < activeSplit {
			timeStr = formatDuration(dispTime)
		}

		// Added \033[K at the end of the line to clear any residual characters
		// from previous prints if a line suddenly gets shorter.
		sb.WriteString(fmt.Sprintf("  %-10s Time: %s%-12s%s (Best: %s%s%s)\033[K\n",
			splitName, segmentColor, timeStr, Reset, Purple, bestStr, Reset))
	}

	sb.WriteString("-----------------------------------------------------\033[K\n")

	if ilMode == 0 {
		totalColor := Cyan
		if bestTotal > 0 && started {
			if currentTotal <= bestTotal {
				totalColor = Green
			} else {
				totalColor = Red
			}
		}

		bestTotalStr := "NONE"
		if bestTotal > 0 {
			bestTotalStr = formatDuration(bestTotal)
		}

		sb.WriteString(fmt.Sprintf("  %-10s Time: %s%-12s%s (Best: %s%s%s)\033[K",
			"TOTAL:", totalColor, formatDuration(currentTotal), Reset, Purple, bestTotalStr, Reset))
	}
	sb.WriteString("\n=====================================================\033[K\n")

	// Status Line Footer
	if !started {
		sb.WriteString(fmt.Sprintf("Status: %sSTOPPED / READY%s", White, Reset))
	} else if paused {
		nextText := fmt.Sprintf("Split %d", activeSplit+2)
		if ilMode > 0 {
			nextText = "Done"
		}
		sb.WriteString(fmt.Sprintf("Status: %sPAUSED (Next WASD starts %s)%s", Yellow, nextText, Reset))
	} else {
		sb.WriteString(fmt.Sprintf("Status: %sRUNNING%s", Green, Reset))
	}

	// Clear any leftover lines below the UI just in case the UI shrunk
	sb.WriteString("\033[J")

	// Print everything to the screen in ONE shot
	fmt.Print(sb.String())
}

var bi bool

func main() {
	// Hide the cursor immediately
	fmt.Print("\033[?25l")

	// Standard cleanup: Ensure the cursor comes back if the user hits Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Print("\033[?25h")
		os.Exit(0)
	}()
	// Parse CLI inputs
	flag.IntVar(&ilMode, "il", 0, "Run a single Individual Level (1-5). Standard mode if 0.")
	flag.Parse()

	// Adjust file paths to prevent overwriting Full Run times during an IL run
	if ilMode < 0 || ilMode > 5 {
		fmt.Println("Error: -il option must be between 1 and 5")
		os.Exit(1)
	}
	if ilMode > 0 {
		filePath = fmt.Sprintf("./times_il_%d", ilMode)
	}

	loadBestTimes()

	var err error
	read, err = IMan.Connect(IMan.ModeBlocking)
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

		if ev.Event.Type == input.EV_KEY {
			switch ev.Event.Code {
			case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
				if ev.Event.Value == 1 { // Key Down
					if ended {
						continue
					}
					if !started {
						started = true
						paused = false

						// If IL Mode is active, offset activeSplit to the selected level
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
						// In IL mode, pausing implies the run is complete, do not advance splits.
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
					// go func() {
					// 	bi = true
					// 	send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 1, Type: input.EV_KEY})
					// 	time.Sleep(20 * time.Millisecond)
					// 	send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 0, Type: input.EV_KEY})
					// 	time.Sleep(20 * time.Millisecond)
					// 	moveMouse(1920/2, (1080/2)+75)
					// 	click()
					// 	bi = false
					// }()
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
						// send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 1, Type: input.EV_KEY})
						// time.Sleep(20 * time.Millisecond)
						// send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 0, Type: input.EV_KEY})
						time.Sleep(20 * time.Millisecond)
						bi = true
						// time.Sleep(2000 * time.Millisecond)
						moveMouse(1920/2, (1080/2)+75)
						click()
						time.Sleep(300 * time.Millisecond)
						moveMouse(596, 223)
						click()
						if ilMode > 0 {
							moveMouse(levelPos[ilMode-1][0], levelPos[ilMode-1][1])
						} else {
							moveMouse(levelPos[0][0], levelPos[0][1])
						}
						click()
						moveMouse(1920/2, 1080/2)
						// time.Sleep(2000 * time.Millisecond)
						bi = false
					}()
					loadBestTimes()
					continue
				}
			case input.KEY_J:
				if ev.Event.Value == 1 {
					go moveMouse(0, 0)
				}
			case input.BTN_RIGHT:
				if ev.Event.Value == 1 { // Click Down
					if started && !paused && !ended {
						paused = true

						finalSegmentTime := accumulatedTime + time.Since(startTime)
						accumulatedTime = finalSegmentTime
						elapsed = finalSegmentTime
						splitTimes[activeSplit] = finalSegmentTime

						// Save PB for the active segment
						if bestTimes[activeSplit] == 0 || finalSegmentTime < bestTimes[activeSplit] {
							bestTimes[activeSplit] = finalSegmentTime
						}

						// End the speedrun if full run hits limit OR if we are in single IL mode
						if ilMode > 0 || activeSplit == totalSplits-1 {
							ended = true
							saveBestTimes()
						}
						// else {
						// 	level += 1
						// 	playLevel(level)
						// }
					}
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

// hl.window_rule({
// 	name = "hjhhkMathasdbreakers",
// 	match = {
// 		class = "^Mathbreakers$",
// 		-- title = "^kitten$",
// 	},
// 	pin = true,
// 	float = true,
// 	no_focus = true,
// 	no_initial_focus = true,
// 	size = { "480.0", "200" },
// 	move = { "0", "30" },
// 	opacity = "1 override",
// 	border_size = 0,
// })

// kitty --class Mathbreakers -c /dev/null go run .

// keyModifier --modify space turbo downFor 20ms delay 20ms --modify space maxPressTime 350ms
