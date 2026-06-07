package main

import "time"

// RunState holds all mutable state for the current speedrun session.
// Keeping state in one struct makes it easy to reset, inspect, and extend.
type RunState struct {
	// Run configuration (set at startup)
	ILMode  int  // 0 = full run, 1-5 = specific IL split
	NoReset bool // if true, ESC does not reset the run

	// Active level tracking
	Level int

	// Timer state
	Started         bool
	Paused          bool
	Ended           bool
	StartTime       time.Time
	AccumulatedTime time.Duration // Time accumulated across pauses in the current split
	Elapsed         time.Duration // Live split time (updated by display ticker)

	// Split tracking
	ActiveSplit int
	SplitTimes  []time.Duration

	// Best-time records (loaded from disk)
	TotalBest        time.Duration
	SOB              time.Duration // Sum of individual-level bests
	WRSob            time.Duration // Sum of world-record bests
	FullRunBestTimes []time.Duration
	ILBestTimes      []time.Duration
	WRTimes          []time.Duration
}

// NewRunState returns an initialized RunState ready to use.
func NewRunState() *RunState {
	return &RunState{
		SplitTimes:       make([]time.Duration, TotalSplits),
		FullRunBestTimes: make([]time.Duration, TotalSplits),
		ILBestTimes:      make([]time.Duration, TotalSplits),
		WRTimes:          make([]time.Duration, TotalSplits),
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
