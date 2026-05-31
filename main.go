package main

import (
	"fmt"
	"log"
	"os/exec"
	"time"

	"github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
)

func main() {
	level := 0
	levelPos := [][]int{
		{546, 293},
		{549, 555},
		{944, 556},
		{547, 808},
		{935, 814},
	}
	im, err := IMan.Connect(IMan.ModeListen)
	print(levelPos[level])
	if err != nil {
		panic(err)
	}
	ticker := time.NewTicker(10 * time.Millisecond)
	done := make(chan bool)
	startTime := time.Now()
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				elapsed := time.Since(startTime)

				// Extract hours, minutes, and seconds from the duration
				h := int(elapsed.Hours())
				m := int(elapsed.Minutes()) % 60
				s := int(elapsed.Seconds()) % 60
				ms := int(elapsed.Milliseconds()) % 1000

				// Format as HH:MM:SS.mmm (padded with leading zeros)
				fmt.Printf("\rTimer: %02d:%02d:%02d.%03d", h, m, s, ms)
			}
		}
	}()
	for {
		ev, err := im.ReadNext()
		if err != nil {
			panic(err)
		}

		switch ev.Event.Code {
		case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
			{
				// print("start")
			}
		}
	}
}
func moveMouse(x, y int) {
	cmd := exec.Command("hyprctl", "dispatch", fmt.Sprintf("hl.dsp.movecursor(%d, %d)", x, y))
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatalf("Command failed with error: %v\nOutput: %s", err, string(output))
	}
	// fmt.Printf("Success:\n%s\n", string(output))
}
