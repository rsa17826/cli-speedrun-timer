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
	"github.com/rsa17826/input-manager/IMan"

	. "speedrun/games"
)

func init() {
	Register(&Game{
		Name:        "vex",
		WindowClass: "vex", // see `hyprctl activewindow`; use WindowTitle instead to match the title
		Splits:      []string{"level"},

		WatchDir:       "./data/vex",
		LevelDoneFile:  "level_cleared.txt",
		LevelResetFile: "levelWasReset",

		StartKeys:   []uint16{input.KEY_W, input.KEY_A, input.KEY_S, input.KEY_D, input.KEY_LEFT, input.KEY_RIGHT, input.KEY_UP, input.KEY_DOWN},
		ResetKeys:   []uint16{input.KEY_TAB},
		RestartKeys: []uint16{input.KEY_ESC},

		Binds: []Bind{},

		PlayLevel: func(c Ctx, split int) {
			c.Send(IMan.WireEvent{Code: input.KEY_LEFTCTRL, Value: 1})
			c.Send(IMan.WireEvent{Code: input.KEY_LEFTSHIFT, Value: 1})
			c.Send(IMan.WireEvent{Code: input.KEY_L, Value: 1})
			c.Sync()
			c.Send(IMan.WireEvent{Code: input.KEY_L, Value: 0})
			c.Send(IMan.WireEvent{Code: input.KEY_LEFTSHIFT, Value: 0})
			c.Send(IMan.WireEvent{Code: input.KEY_LEFTCTRL, Value: 0})
			c.Sync()
			// // Launch split `split` (0-indexed): write a file, click, send keys...
			// c.MoveMouse(500, 300)
			// c.Click()
		},
	})
}
