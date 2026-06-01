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
var elapsed time.Duration

func pt() {
	// Extract hours, minutes, and seconds from the duration
	h := int(elapsed.Hours())
	m := int(elapsed.Minutes()) % 60
	s := int(elapsed.Seconds()) % 60
	ms := int(elapsed.Milliseconds()) % 1000

	// Format as HH:MM:SS.mmm (padded with leading zeros)
	if !started {
		fmt.Printf("\nTimer: %02d:%02d:%02d.%03d, STOPPED", h, m, s, ms)
	} else if paused {
		fmt.Printf("\nTimer: %02d:%02d:%02d.%03d, PAUSED", h, m, s, ms)
	} else {
		fmt.Printf("\nTimer: %02d:%02d:%02d.%03d", h, m, s, ms)
	}
}

func main() {
	var err error
	read, err = IMan.Connect(IMan.ModeBlocking)
	send, err = IMan.Connect(IMan.ModeInjection)
	// read, err := IMan.Connect(IMan.ModeListen, IMan.ModeInjection)
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
				elapsed = time.Since(startTime)
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
				{
					if paused {
						paused = false
					} else {
						if !started {
							started = true
							startTime = time.Now()
						}
					}
				}
			case input.KEY_ESC:
				{
					started = false
				}
			case input.BTN_RIGHT:
				{
					paused = true
				}
				// case input.BTN_RIGHT:
				// 	{
				// 		if ev.Event.Value == 1 {
				// 			playLevel(1)
				// 			block = 1
				// 		}
				// 	}
				// case input.KEY_KP7:
				// 	{
				// 		read.BlockInput(1)
				// 		if ev.Event.Value == 1 {
				// 			println("7")
				// 			go playLevel(0)
				// 		}
				// 		continue
				// 	}
				// case input.KEY_ESC:
				// 	{
				// 		read.BlockInput(0)
				// 		if ev.Event.Value == 1 {
				// 			println("a")
				// 			go func() {
				// 				send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_X, Value: -99999})
				// 				send.Send(IMan.WireEvent{})
				// 				send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_Y, Value: -999})
				// 				send.Send(IMan.WireEvent{})
				// 				moveMouse(947, 621)
				// 				click()
				// 			}()
				// 		}
				// 		continue
				// 	}
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
	// playbtn
	// send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_X, Value: 1})

	moveMouse(0, 0)
	moveMouse(596, 223)
	click()
	moveMouse(0, 0)
	moveMouse(levelPos[level][0], levelPos[level][1])
	click()
	moveMouse(1920/2, 1080/2)
}
func click() {
	// 1. Mouse Down
	send.Send(IMan.WireEvent{
		Type:  input.EV_KEY,
		Code:  input.BTN_LEFT,
		Value: 1,
	})
	send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)

	// 2. Mouse Up
	send.Send(IMan.WireEvent{
		Type:  input.EV_KEY,
		Code:  input.BTN_LEFT,
		Value: 0,
	})
	send.Send(IMan.WireEvent{})
	time.Sleep(150 * time.Millisecond)
}
