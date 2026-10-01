package main

import (
	"os"
	"path"
	"strconv"
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
)

// moveMouse moves the hardware mouse cursor to absolute screen coordinates
// by first resetting to (0,0) then moving in two steps for precision.
func (a *App) sendAbs(code uint16, value int32) {
	if err := a.send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: code, Value: value}); err != nil {
		panic(err)
	}
}

const layoutW, layoutH = 4560, 4142
const monX, monY = 2000, 28 // test_bottom origin

func (a *App) moveMouse(_x, _y int32) {
	x := ((monX+_x)*32768 + layoutW - 1) / layoutW
	y := ((monY+_y)*32768 + layoutH - 1) / layoutH

	// The kernel drops an ABS event equal to the axis's last value, and real-mouse
	// movement doesn't update that value, so repeating a target would be ignored.
	// A different value first guarantees the target registers.
	nx, ny := x-1, y-1
	if x == 0 {
		nx = 1
	}
	if y == 0 {
		ny = 1
	}
	a.sendAbs(input.ABS_X, nx)
	a.sendAbs(input.ABS_Y, ny)
	a.sendAbs(input.ABS_X, x)
	a.sendAbs(input.ABS_Y, y)
	a.sendSync()
}

// click sends a left mouse button press and release.
func (a *App) click() {
	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 1})
	a.send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)

	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 0})
	a.send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)
}

// playLevel navigates to and launches the given level index.
func (a *App) playLevel(i int) {
	// a.blocking = true
	a.rs.Level = i
	os.WriteFile(path.Join(WatchDir, "mode"), []byte(strconv.Itoa(i)), 0644)
	// a.clickPlayButton()
	// a.moveMouse(LevelPositions[i][0], LevelPositions[i][1])
	// a.click()
	// a.blocking = false
}

// clickExitLevelButton clicks the in-level exit/return button.
func (a *App) clickExitLevelButton() {
	a.moveMouse(ScreenWidth/2, ScreenHeight/2+75)
	a.click()
}

// clickPlayButton clicks the main play/level-select button.
func (a *App) clickPlayButton() {
	a.moveMouse(596, 223)
	a.click()
}

// sendRel sends a relative axis movement event.

// sendSync sends a sync event to flush pending input events.
func (a *App) sendSync() {
	if err := a.send.Send(IMan.WireEvent{}); err != nil {
		println(err)
	}
}
