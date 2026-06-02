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
var totalBest time.Duration
var sob time.Duration

// Split Variables
const totalSplits = 5

// Separate persistent files for distinct category matching
var fullRunPath = "./times_full_run"
var ilPath = "./times_il_all"

var activeSplit = 0
var splitTimes = make([]time.Duration, totalSplits)

// Distinct tracking arrays to print both contexts concurrently
var fullRunBestTimes = make([]time.Duration, totalSplits)
var ilBestTimes = make([]time.Duration, totalSplits)

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

// Load records across both configurations
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

	// Repeat the padding character to fill the spaces
	leftStr := strings.Repeat(padChar, leftPadding)
	rightStr := strings.Repeat(padChar, rightPadding)

	if bn {
		return leftStr + text + rightStr + "\n"
	}
	return leftStr + text + rightStr
}
func pt() {
	var size int = 83
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

		// Selection logic for target delta color indicators
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

		timeStr := "--:--:--.---"
		if started && i <= activeSplit || i < activeSplit {
			timeStr = formatDuration(dispTime)
		}

		p := int((float64(ilBestTimes[i]) / float64(fullRunBestTimes[i])) * 100.0)
		pc := Yellow
		if p == 100 {
			pc = Green
		}
		fmt.Fprintf(&sb, "  %-10s Time: %s%-12s%s (Run Best: %s%-12s%s IL Best: %s%s%s) %s%d%%%s\033[K\n",
			splitName, segmentColor, timeStr, Reset, Purple, fullBestStr, Reset, Purple, ilBestStr, Reset, pc, p, Reset)
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
		fmt.Fprintf(&sb, "  %-10s Time: %s%-12s%s (Best Total: %s%s%s) (SOB: %s%s%s) %s%d%%%s\033[K\n",
			"TOTAL:", totalColor, formatDuration(currentTotal), Reset, Purple, bestTotalStr, Reset, Purple, formatDuration(sob), Reset, pc, p, Reset)
	}
	sb.WriteString(pad("", "=", size, false))
	sb.WriteString("\033[K\n")

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

	sb.WriteString("\033[J")
	fmt.Print(sb.String())
}

var bi bool

func main() {
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

						// Evaluate and record IL performance regardless of current mode context
						if ilBestTimes[activeSplit] == 0 || finalSegmentTime < ilBestTimes[activeSplit] {
							ilBestTimes[activeSplit] = finalSegmentTime
							saveFile("il") // Updates IL tracking file immediately on segment completion
						}

						if ilMode > 0 {
							ended = true
						} else if activeSplit == totalSplits-1 {
							// Full Run Complete calculation
							var totalRunTime time.Duration
							for _, t := range splitTimes {
								totalRunTime += t
							}

							// Save Full Run records if overall benchmark is surpassed
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
								time.Sleep(300 * time.Millisecond)
								moveMouse(596, 223)
								click()
								moveMouse(levelPos[activeSplit+1][0], levelPos[activeSplit+1][1])
								click()
								moveMouse(1920/2, 1080/2)
							}()
						}
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
