// The "_template" folder is ignored by the Go tool.
// To add a game: copy this folder to games/yourgame, rename the package to
// match, then add one line to games_import.go:
//
//	_ "speedrun/games/yourgame"
//
// and run:  go run . yourgame
package yourgame

import (
	input "github.com/rsa17826/go-input-lib"

	. "speedrun/games"
)

func init() {
	Register(&Game{
		Name:        "blinx",
		WindowClass: "app.xemu.xemu",
		Splits:      []string{"level 1-1", "level 1-2", "level 1-3", "level 1-4", "level 2-1", "level 2-2", "level 2-3", "level 2-4", "level 3-1", "level 3-2", "level 3-3", "level 3-4", "level 4-1", "level 4-2", "level 4-3", "level 4-4", "level 5-1", "level 5-2", "level 5-3", "level 5-4", "level 6-1", "level 6-2", "level 6-3", "level 6-4", "level 7-1", "level 7-2", "level 7-3", "level 7-4", "level 8-1", "level 8-2", "level 8-3", "level 8-4", "level 9-1"},

		WatchDir:       "./data/blinx",
		LevelDoneFile:  "level_cleared.txt",
		LevelResetFile: "levelWasReset",

		StartKeys:   []uint16{input.KEY_ENTER},
		ResetKeys:   []uint16{input.KEY_ENTER},
		RestartKeys: []uint16{input.KEY_ESC},

		Binds: []Bind{
			// {
			// 	Key:   input.KEY_ENTER,
			// 	Block: true,
			// 	OnPress: func(c Ctx) {
			// 		c.TapKey(input.KEY_R, 40*time.Millisecond)
			// 	},
			// },
		},

		PlayLevel: func(c Ctx, split int) {
			// Launch split `split` (0-indexed): write a file, click, send keys...
			c.MoveMouse(500, 300)
			c.Click()
		},
	})
}
