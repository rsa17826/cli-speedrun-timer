package main

import (
	"fmt"
	"image/color"
	"log"
	"os/exec"
	"time"

	"fyne.io/fyne/app"
	"fyne.io/fyne/canvas"
	"fyne.io/fyne/v2"
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
	go func() {
		myApp := app.New()
		myWindow := myApp.NewWindow("Timer")

		// Create a large, bold text element for the timer
		timerText := canvas.NewText("0s", color.White)
		timerText.TextSize = 48
		timerText.Alignment = fyne.TextAlignCenter

		// Set the text as the window content and set window size
		myWindow.SetContent(timerText)
		myWindow.Resize(fyne.NewSize(250, 150))

		// Run the background timer loop in a separate goroutine
		go func() {
			ticker := time.NewTicker(time.Second)
			defer ticker.Stop()

			seconds := 0
			for range ticker.C {
				seconds++
				timerText.Text = fmt.Sprintf("%ds", seconds)
				timerText.Refresh() // Tells Fyne to redraw the text element
			}
		}()

		// Show the window and block until the app is closed
		myWindow.ShowAndRun()
	}()
	for {
		ev, err := im.ReadNext()
		if err != nil {
			panic(err)
		}
		switch ev.Event.Code {
		case input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D:
			{
				print("start")
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
