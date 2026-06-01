package main

import (
	"fmt"
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
var startTime time.Time
var accumulatedTime time.Duration // Tracks time accumulated across pauses for current split
var elapsed time.Duration         // Tracks current live split time

// Split Variables
const totalSplits = 5

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
)

func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

func pt() {
	// Clear screen to redraw the full layout cleanly without trailing text
	fmt.Print("\033[H\033[2J")

	fmt.Println("============ SPEEDRUN SPLITS ============")
	for i := 0; i < totalSplits; i++ {
		splitName := fmt.Sprintf("Segment %d:", i+1)
		bestStr := "NONE"
		if bestTimes[i] > 0 {
			bestStr = formatDuration(bestTimes[i])
		}

		if started && i == activeSplit {
			// Highlights the currently active split
			statusColor := Green
			if paused {
				statusColor = Yellow
			}
			fmt.Printf("%s %-10s Current: %s%s%s (Best: %s%s%s)\n",
				statusColor, splitName, statusColor, formatDuration(elapsed), Reset, Purple, bestStr, Reset)
		} else {
			// Shows finished splits or upcoming splits
			recordedStr := "--:--:--.--"
			if splitTimes[i] > 0 {
				recordedStr = formatDuration(splitTimes[i])
			}
			fmt.Printf("  %-10s Time: %s%-12s%s (Best: %s%s%s)\n",
				splitName, Cyan, recordedStr, Reset, Purple, bestStr, Reset)
		}
	}
	fmt.Println("=========================================")

	// Status Line Footer
	if !started {
		fmt.Printf("Status: %sSTOPPED / READY%s", Red, Reset)
	} else if paused {
		fmt.Printf("Status: %sPAUSED (Next WASD starts Split %d)%s", Yellow, activeSplit+1, Reset)
	} else {
		fmt.Printf("Status: %sRUNNING%s", Green, Reset)
	}
}

func main() {
	var err error
	read, err = IMan.Connect(IMan.ModeBlocking)
	send, err = IMan.Connect(IMan.ModeInjection)
	if err != nil {
		panic(err)
	}

	ticker := time.NewTicker(10 * time.Millisecond)
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

				// Live update of current split
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
					if !started {
						// Hard Fresh Start from zero
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
						// Unpausing: Move to NEXT split segment if we haven't maxed out
						if activeSplit < totalSplits-1 {
							activeSplit++
							elapsed = 0
							accumulatedTime = 0
							paused = false
							startTime = time.Now()
						} else {
							// If we completed all 5 segments, unpausing just resumes final segment time tracking
							paused = false
							startTime = time.Now()
						}
					}
				}
			case input.KEY_ESC:
				if ev.Event.Value == 1 {
					started = false
					paused = false
					activeSplit = 0
					accumulatedTime = 0
					elapsed = 0
					// Reset temporary session run times, keeping PB records intact
					for i := range splitTimes {
						splitTimes[i] = 0
					}
				}
			case input.BTN_RIGHT:
				if ev.Event.Value == 1 { // Click Down
					if started && !paused {
						paused = true

						// Commit current accumulated run segment time
						finalSegmentTime := accumulatedTime + time.Since(startTime)
						accumulatedTime = finalSegmentTime
						elapsed = finalSegmentTime
						splitTimes[activeSplit] = finalSegmentTime

						// Check and update Personal Best time for this specific segment
						if bestTimes[activeSplit] == 0 || finalSegmentTime < bestTimes[activeSplit] {
							bestTimes[activeSplit] = finalSegmentTime
						}
					}
				}
			}
		}
		read.BlockInput(0)
	}
}

func moveMouse(x, y int32) {
	send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_X, Value: x})
	send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_Y, Value: y})
	send.Send(IMan.WireEvent{})
	time.Sleep(150 * time.Millisecond)
}

func playLevel(i int) {
	level = i
	moveMouse(0, 0)
	moveMouse(596, 223)
	click()
	moveMouse(0, 0)
	moveMouse(levelPos[level][0], levelPos[level][1])
	click()
	moveMouse(1920/2, 1080/2)
}

func click() {
	send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 1})
	send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)

	send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 0})
	send.Send(IMan.WireEvent{})
	time.Sleep(150 * time.Millisecond)
}
