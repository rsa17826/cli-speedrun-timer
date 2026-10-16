// Package mathbreakers is the Mathbreakers game definition.
// Its times_* files and windowrules.lua live in this folder.
package mathbreakers

import (
	"os"
	"path"
	"strconv"
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"

	. "speedrun/games"
)

const mathbreakersWatchDir = "/home/nyix/projects/mathbreakers-ghost-speedrun-mod/mathbreakers"

// mathbreakersLevelPositions maps a level index to its click coordinates
// (only used by the old click-based navigation, kept for reference).
var mathbreakersLevelPositions = [][2]int32{
	{546, 293},
	{549, 555},
	{944, 556},
	{547, 808},
	{935, 814},
}

func init() {
	Register(&Game{
		Name:        "mathbreakers",
		WindowClass: "mathbreakers.exe",
		Splits:      []string{"Level 1", "Level 2", "Level 3", "Level 4", "Level 5"},

		DataDir: "./data/mathbreakers",

		WatchDir:       mathbreakersWatchDir,
		LevelDoneFile:  "level_cleared.txt",
		LevelResetFile: "levelWasReset",

		StartKeys:   []uint16{input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D},
		ResetKeys:   []uint16{input.KEY_TAB},
		RestartKeys: []uint16{input.KEY_ESC},

		Binds: []Bind{
			{
				// Right click: scroll on the last split, otherwise tap R.
				Key:   input.BTN_RIGHT,
				Block: true,
				OnPress: func(c Ctx) {
					if c.ActiveSplit() == 4 {
						c.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_WHEEL, Value: 1})
						return
					}
					c.TapKey(input.KEY_R, 40*time.Millisecond)
				},
			},
		},

		PlayLevel: func(c Ctx, split int) {
			err := os.WriteFile(path.Join(mathbreakersWatchDir, "mode"), []byte(strconv.Itoa(split)), 0644)
			if err != nil {
				panic(err)
			}
		},
	})
}
