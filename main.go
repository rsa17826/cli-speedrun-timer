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
var accumulatedTime time.Duration // Tracks time accumulated across pauses
var elapsed time.Duration         // Tracks time accumulated across pauses

func pt() {
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	s := int(elapsed.Seconds()) % 60
	ms := int(elapsed.Milliseconds()) % 1000

	if !started {
		fmt.Printf("\rTimer: %02d:%02d:%02d.%03d, STOPPED", h, m, s, ms)
	} else if paused {
		fmt.Printf("\rTimer: %02d:%02d:%02d.%03d, PAUSED ", h, m, s, ms)
	} else {
		fmt.Printf("\rTimer: %02d:%02d:%02d.%03d         ", h, m, s, ms)
	}
}
func main() {
	var err error
	read, err = IMan.Connect(IMan.ModeBlocking)
	send, err = IMan.Connect(IMan.ModeInjection)
	if err != nil {
		panic(err)
	}

	// Keep a single 10ms ticker for the UI/terminal updates
	ticker := time.NewTicker(10 * time.Millisecond)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				// 1. If it's not started, or it's paused, we don't need to spam updates.
				// We print the static state once and wait.
				if !started || paused {
					pt()

					// To prevent this loop from spinning aggressively while paused,
					// we just let the 10ms ticker tick, but you could also implement
					// a condition variable or channel here if you want 0% CPU usage on pause.
					continue
				}

				// 2. If actively running, update the elapsed time dynamically.
				// time.Since() provides nanosecond precision regardless of the 10ms loop speed.
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

		// Only trigger on key down/up event values depending on your needs,
		// but checking key type here:
		if ev.Event.Type == input.EV_KEY {
			switch ev.Event.Code {
			case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
				if ev.Event.Value == 1 { // Only trigger on Key Down
					if !started {
						started = true
						paused = false
						startTime = time.Now()
						accumulatedTime = 0
					} else if paused {
						// Unpausing: reset the baseline startTime to now
						paused = false
						startTime = time.Now()
					}
				}
			case input.KEY_ESC:
				if ev.Event.Value == 1 {
					started = false
					paused = false
					accumulatedTime = 0
				}
			case input.BTN_RIGHT:
				if ev.Event.Value == 1 { // Only trigger on Click Down
					if started && !paused {
						paused = true
						// Save the chunk of time passed since the last unpause/start
						accumulatedTime += time.Since(startTime)
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
