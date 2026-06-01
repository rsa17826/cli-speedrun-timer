package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
const filePath = "./times"

var activeSplit = 0
var splitTimes = make([]time.Duration, totalSplits)
var bestTimes = make([]time.Duration, totalSplits)

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
	i := 0
	for scanner.Scan() && i < totalSplits {
		ms, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err == nil {
			bestTimes[i] = time.Duration(ms) * time.Millisecond
		}
		i++
	}
}

// Save best times to file
func saveBestTimes() {
	file, err := os.Create(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	for _, t := range bestTimes {
		fmt.Fprintf(file, "%d\n", t.Milliseconds())
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
	// Clear screen to redraw cleanly
	fmt.Print("\033[H\033[2J")

	fmt.Println("================== SPEEDRUN SPLITS ==================")

	var currentTotal time.Duration
	var bestTotal time.Duration

	for i := 0; i < totalSplits; i++ {
		splitName := fmt.Sprintf("Segment %d:", i+1)
		bestTotal += bestTimes[i]

		var dispTime time.Duration
		var segmentColor = Cyan

		if i < activeSplit {
			dispTime = splitTimes[i]
			currentTotal += splitTimes[i]
			// Completed segment: color based on whether it beat PB
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
			// Active segment live color profiling
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

		fmt.Printf("  %-10s Time: %s%-12s%s (Best: %s%s%s)\n",
			splitName, segmentColor, timeStr, Reset, Purple, bestStr, Reset)
	}

	fmt.Println("-----------------------------------------------------")

	// Determine Total Time Pace Color
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

	fmt.Printf("  %-10s Time: %s%-12s%s (Best: %s%s%s)",
		"TOTAL:", totalColor, formatDuration(currentTotal), Reset, Purple, bestTotalStr, Reset)
	fmt.Println("\n=====================================================")

	// Status Line Footer
	if !started {
		fmt.Printf("Status: %sSTOPPED / READY%s", White, Reset)
	} else if paused {
		fmt.Printf("Status: %sPAUSED (Next WASD starts Split %d)%s", Yellow, activeSplit+2, Reset)
	} else {
		fmt.Printf("Status: %sRUNNING%s", Green, Reset)
	}
}

var bi bool

func main() {
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
						activeSplit = 0
						elapsed = 0
						accumulatedTime = 0
						for i := range splitTimes {
							splitTimes[i] = 0
						}
						startTime = time.Now()
					} else if paused {
						activeSplit++
						elapsed = 0
						accumulatedTime = 0
						paused = false
						startTime = time.Now()
					}
				}
			case input.KEY_TAB:
				{
					if ev.Event.Value == 1 {
						go func() {
							bi = true
							send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 1, Type: input.EV_KEY})
							time.Sleep(20 * time.Millisecond)
							send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 0, Type: input.EV_KEY})
							time.Sleep(20 * time.Millisecond)
							// time.Sleep(2000 * time.Millisecond)
							moveMouse(1920/2, (1080/2)+75)
							click()
							// time.Sleep(2000 * time.Millisecond)
							bi = false
						}()
					}
				}
			case input.KEY_ESC:
				if ev.Event.Value == 1 {
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
					// go func() {
					// 	send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 1, Type: input.EV_KEY})
					// 	time.Sleep(20 * time.Millisecond)
					// 	send.Send(IMan.WireEvent{Code: input.KEY_ESC, Value: 0, Type: input.EV_KEY})
					// 	time.Sleep(20 * time.Millisecond)
					// 	bi = true
					// 	// time.Sleep(2000 * time.Millisecond)
					// 	moveMouse(1920/2, (1080/2)+75)
					// 	click()
					// 	// time.Sleep(2000 * time.Millisecond)
					// 	bi = false
					// }()
					loadBestTimes()
					continue
				}
			case input.KEY_J:
				{
					if ev.Event.Value == 1 { // Click Down
						go moveMouse(0, 0)
					}
				}
			case input.BTN_RIGHT:
				if ev.Event.Value == 1 { // Click Down
					if started && !paused && !ended {
						paused = true

						finalSegmentTime := accumulatedTime + time.Since(startTime)
						accumulatedTime = finalSegmentTime
						elapsed = finalSegmentTime
						splitTimes[activeSplit] = finalSegmentTime

						// Save if it's a personal record for this segment
						if bestTimes[activeSplit] == 0 || finalSegmentTime < bestTimes[activeSplit] {
							bestTimes[activeSplit] = finalSegmentTime
						}
						if activeSplit == totalSplits-1 {
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
