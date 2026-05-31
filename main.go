package main

import (
	"log"
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

func main() {
	var err error
	read, err = IMan.Connect(IMan.ModeBlocking)
	send, err = IMan.Connect(IMan.ModeInjection)
	// read, err := IMan.Connect(IMan.ModeListen, IMan.ModeInjection)
	print(levelPos[level])
	if err != nil {
		panic(err)
	}
	// ticker := time.NewTicker(1 * time.Millisecond)
	// done := make(chan bool)
	// startTime := time.Now()
	// go func() {
	// 	for {
	// 		select {
	// 		case <-done:
	// 			return
	// 		case <-ticker.C:
	// 			elapsed := time.Since(startTime)

	// 			// Extract hours, minutes, and seconds from the duration
	// 			h := int(elapsed.Hours())
	// 			m := int(elapsed.Minutes()) % 60
	// 			s := int(elapsed.Seconds()) % 60
	// 			ms := int(elapsed.Milliseconds()) % 1000

	// 			// Format as HH:MM:SS.mmm (padded with leading zeros)
	// 			fmt.Printf("\nTimer: %02d:%02d:%02d.%03d", h, m, s, ms)
	// 		}
	// 	}
	// }()
	for {
		ev, err := read.ReadNext()
		if err != nil {
			panic(err)
		}
		var block uint8 = 0

		switch ev.Event.Code {
		case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
			{
				println("start")
			}
		// case input.BTN_RIGHT:
		// 	{
		// 		if ev.Event.Value == 1 {
		// 			playLevel(1)
		// 			block = 1
		// 		}
		// 	}
		case input.KEY_KP7:
			{
				if ev.Event.Value == 1 {
					playLevel(1)
					block = 1
				}
			}
		}
		read.BlockInput(block)
	}
}
func moveMouse(x, y int32) {
	err := send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_X, Value: x})
	if err != nil {
		log.Printf("Error sending movement: %v", err)
	}
	err = send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_Y, Value: y})

	if err != nil {
		log.Printf("Error sending movement: %v", err)
	}

	err = send.Send(IMan.WireEvent{})
	if err != nil {
		log.Printf("Error sending sync: %v", err)
	}
	time.Sleep(150 * time.Millisecond)
}

func playLevel(i int) {
	level = i
	// playbtn
	send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: input.ABS_X, Value: 1})

	moveMouse(596, 223)
	click()
	moveMouse(levelPos[level][0], levelPos[level][1])
	// moveMouse(596, 223)
	// time.Sleep(150 * time.Millisecond)
	// send.Send(IMan.WireEvent{
	// 	// Type: input,
	// 	Value: 1,
	// 	Code:  input.BTN_LEFT,
	// })
	// send.Send(IMan.WireEvent{})
	// time.Sleep(150 * time.Millisecond)
	// send.Send(IMan.WireEvent{
	// 	Type: input.EV_REL,
	// 	// Type:  input.EV_ABS,
	// 	Value: -9,
	// 	Code:  input.REL_X,
	// })
	// send.Send(IMan.WireEvent{})
	// send.Send(IMan.WireEvent{
	// 	// Type: input,
	// 	Value: 0,
	// 	Code:  input.BTN_LEFT,
	// })
	// send.Send(IMan.WireEvent{})
	// time.Sleep(150 * time.Millisecond)
}
func click() {
	// 1. Mouse Down
	send.Send(IMan.WireEvent{
		Type:  input.EV_KEY, // Explicitly state this is a key/button event
		Code:  input.BTN_LEFT,
		Value: 1, // 1 = Press
	})
	send.Send(IMan.WireEvent{Type: input.EV_SYN}) // Explicit Sync
	time.Sleep(150 * time.Millisecond)

	// 2. Mouse Up
	send.Send(IMan.WireEvent{
		Type:  input.EV_KEY,
		Code:  input.BTN_LEFT,
		Value: 0, // 0 = Release
	})
	send.Send(IMan.WireEvent{Type: input.EV_SYN}) // Explicit Sync
	time.Sleep(150 * time.Millisecond)
}
