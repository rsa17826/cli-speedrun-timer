// The "_template" folder is ignored by the Go tool.
// To add a game: copy this folder to games/yourgame, rename the package to
// match, then add one line to games_import.go:
//
//	_ "speedrun/games/yourgame"
//
// and run:  go run . yourgame
package yourgame

import (
	"time"

	input "github.com/rsa17826/go-input-lib"

	. "speedrun/games"
)

func init() {
	Register(&Game{
		Name:        "yourgame",
		WindowClass: "yourgame.exe", // see `hyprctl activewindow`; use WindowTitle instead to match the title
		Splits:      []string{"Level 1", "Level 2", "Level 3"},

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
