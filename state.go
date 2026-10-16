package main

import (
	"time"

	"speedrun/games"
)

// RunState holds all mutable state for the current speedrun session.
type RunState struct {
	Game  *games.Game
	Total int // number of splits, len(Game.Splits)

	// Run configuration (set at startup)
	ILMode  int  // 0 = full run, 1-Total = specific IL split
	NoReset bool // if true, ESC does not reset the run

	// Active level tracking
	Level int

	// Timer state
	Started         bool
	Paused          bool
	Ended           bool
	StartTime       time.Time
	AccumulatedTime time.Duration
	Elapsed         time.Duration

	// Split tracking
	ActiveSplit int
	SplitTimes  []time.Duration

	// Best-time records (loaded from disk)
	TotalBest        time.Duration
	SOB              time.Duration
	WRSob            time.Duration
	FGWR             time.Duration
	FullRunBestTimes []time.Duration
	ILBestTimes      []time.Duration
	WRTimes          []time.Duration
}

// NewRunState returns an initialized RunState for the given game.
func NewRunState(g *games.Game) *RunState {
	n := len(g.Splits)
	return &RunState{
		Game:             g,
		Total:            n,
		SplitTimes:       make([]time.Duration, n),
		FullRunBestTimes: make([]time.Duration, n),
		ILBestTimes:      make([]time.Duration, n),
		WRTimes:          make([]time.Duration, n),
	}
}

// Reset clears all active run state while leaving best times intact.
func (rs *RunState) Reset() {
	rs.Started = false
	rs.Paused = false
	rs.Ended = false
	rs.ActiveSplit = 0
	rs.AccumulatedTime = 0
	rs.Elapsed = 0
	for i := range rs.SplitTimes {
		rs.SplitTimes[i] = 0
	}
}
