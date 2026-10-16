// Files starting with "_" are ignored by the Go tool.
// Copy this to games/yourgame.go (and delete this comment) to add a game,
// then run:  go run . yourgame

package games

import (
	"time"

	input "github.com/rsa17826/go-input-lib"
)

// "gTree (|1|2)[0-9]",
func init() {
	Register(&Game{
		Name:        "mathquest",
		WindowClass: "helium",

		Splits: []string{"gTree 0", "gTree 1", "gTree 2", "gTree 3", "gTree 4", "gTree 5", "gTree 6", "gTree 7", "gTree 8", "gTree 9", "gTree 10", "gTree 11", "gTree 12", "gTree 13", "gTree 14", "gTree 15", "gTree 16", "gTree 17", "gTree 18", "gTree 19", "gTree 20", "gTree 21", "gTree 22", "gTree 23", "gTree 24", "gTree 25"},

		DataDir: "./data/mathquest", // must exist; holds times_full_run, times_il_all, wrs

		WatchDir:       "./data/mathquest",
		LevelDoneFile:  "split",
		LevelResetFile: "nyil",

		StartKeys:   []uint16{input.KEY_LEFT, input.KEY_UP, input.KEY_DOWN, input.KEY_RIGHT},
		ResetKeys:   []uint16{input.KEY_TAB},
		RestartKeys: []uint16{input.KEY_F5},

		Binds: []Bind{
			{
				Key:   input.KEY_F5,
				Block: false,
				OnPress: func(c Ctx) {
					c.TapKey(input.KEY_F5, 40*time.Millisecond)
				},
			},
		},

		PlayLevel: func(c Ctx, split int) {
			c.TapKey(input.KEY_F5, 40*time.Millisecond)
			// // Launch split `split` (0-indexed): write a file, click, send keys...
			// c.MoveMouse(500, 300)
			// c.Click()
		},
	})
}
