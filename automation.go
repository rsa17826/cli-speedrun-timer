package main

import (
	"time"

	input "github.com/rsa17826/go-input-lib"
	"github.com/rsa17826/input-manager/IMan"
)

// Machine-specific monitor layout used to translate coordinates to absolute
// ABS values.
const layoutW, layoutH = 4560, 4142
const monX, monY = 2000, 28 // test_bottom origin

// Send implements games.Ctx.
func (a *App) Send(ev IMan.WireEvent) error {
	return a.send.Send(ev)
}

func (a *App) sendAbs(code uint16, value int32) {
	if err := a.send.Send(IMan.WireEvent{Type: input.EV_ABS, Code: code, Value: value}); err != nil {
		panic(err)
	}
}

// MoveMouse implements games.Ctx.
func (a *App) MoveMouse(_x, _y int32) {
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
	a.Sync()
}

// Click implements games.Ctx.
func (a *App) Click() {
	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 1})
	a.send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)

	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: input.BTN_LEFT, Value: 0})
	a.send.Send(IMan.WireEvent{})
	time.Sleep(30 * time.Millisecond)
}

// TapKey implements games.Ctx.
func (a *App) TapKey(code uint16, hold time.Duration) {
	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: code, Value: 1})
	time.Sleep(hold)
	a.send.Send(IMan.WireEvent{Type: input.EV_KEY, Code: code, Value: 0})
}

// Scroll implements games.Ctx.
func (a *App) Scroll(amount int32) {
	a.send.Send(IMan.WireEvent{Type: input.EV_REL, Code: input.REL_WHEEL, Value: amount})
}

// ActiveSplit implements games.Ctx.
func (a *App) ActiveSplit() int {
	return a.rs.ActiveSplit
}

// Sync implements games.Ctx.
func (a *App) Sync() {
	if err := a.send.Send(IMan.WireEvent{}); err != nil {
		println(err)
	}
}

// playLevel records the level and lets the game launch it.
func (a *App) playLevel(i int) {
	a.rs.Level = i
	a.rs.Game.PlayLevel(a, i)
}
