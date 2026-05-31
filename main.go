package main

import (
	"fmt"
	"log"
	"os/exec"

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
