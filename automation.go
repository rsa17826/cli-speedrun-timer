package main

import (
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
)

// moveMouse moves the hardware mouse cursor to absolute screen coordinates
// by first resetting to (0,0) then moving in two steps for precision.
func (a *App) moveMouse(x, y int32) {
	// Reset to top-left corner.
	a.sendRel(input.REL_X, -9999)
	a.sendSync()
	a.sendRel(input.REL_Y, -9999)
	a.sendSync()
	time.Sleep(10 * time.Millisecond)

	a.sendRel(input.REL_X, -10000)
	a.sendSync()
	a.sendRel(input.REL_Y, -10000)
	a.sendSync()
	time.Sleep(10 * time.Millisecond)

	// Move to target position.
	a.sendRel(input.REL_X, x/2)
	a.sendSync()
	a.sendRel(input.REL_Y, y/2)
	a.sendSync()
	time.Sleep(10 * time.Millisecond)
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
	a.blocking = true
	a.rs.Level = i
	a.clickPlayButton()
	a.moveMouse(LevelPositions[i][0], LevelPositions[i][1])
	a.click()
	a.blocking = false
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
func (a *App) sendRel(code uint16, value int32) {
	if err := a.send.Send(IMan.WireEvent{Type: input.EV_REL, Code: code, Value: value}); err != nil {
		println(err)
	}
}

// sendSync sends a sync event to flush pending input events.
func (a *App) sendSync() {
	if err := a.send.Send(IMan.WireEvent{}); err != nil {
		println(err)
	}
}
