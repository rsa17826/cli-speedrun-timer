// Files starting with "_" are ignored by the Go tool.
// Copy this to games/yourgame.go (and delete this comment) to add a game,
// then run:  go run . yourgame

package games

import (
	"time"

	input "github.com/rsa17826/go-input-lib"
)

func init() {
	Register(&Game{
		Name:        "yourgame",
		WindowClass: "yourgame.exe", // see `hyprctl activewindow`
		Splits:      []string{"Level 1", "Level 2", "Level 3"},

		DataDir: "./data/yourgame", // must exist; holds times_full_run, times_il_all, wrs

		WatchDir:       "/path/the/game/writes/into",
		LevelDoneFile:  "level_cleared.txt",
		LevelResetFile: "levelWasReset",

		StartKeys:   []uint16{input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D},
		ResetKeys:   []uint16{input.KEY_TAB},
		RestartKeys: []uint16{input.KEY_ESC},

		Binds: []Bind{
			{
				Key:   input.BTN_RIGHT,
				Block: true,
				OnPress: func(c Ctx) {
					c.TapKey(input.KEY_R, 40*time.Millisecond)
				},
			},
		},

		PlayLevel: func(c Ctx, split int) {
			// Launch split `split` (0-indexed): write a file, click, send keys...
			c.MoveMouse(500, 300)
			c.Click()
		},
	})
}
