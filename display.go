package main

import (
	"fmt"
	"strings"
	"time"
)

// Render prints the full split display to the terminal, using ANSI codes
// to redraw in place without flickering.
func Render(rs *RunState) {
	var sb strings.Builder
	sb.WriteString("\033[H") // move cursor to top-left

	if rs.ILMode > 0 {
		sb.WriteString(pad(fmt.Sprintf(" INDIVIDUAL LEVEL (IL %d) ", rs.ILMode), "=", DisplayWidth, true))
	} else {
		sb.WriteString(pad(" SPEEDRUN SPLITS ", "=", DisplayWidth, true))
	}

	var currentTotal time.Duration

	for i := range TotalSplits {
		if rs.ILMode > 0 && i != rs.ILMode-1 {
			continue
		}

		splitName := fmt.Sprintf("Segment %d:", i+1)
		var dispTime time.Duration
		segmentColor := Cyan

		targetTime := rs.FullRunBestTimes[i]
		if rs.ILMode > 0 {
			targetTime = rs.ILBestTimes[i]
		}

		if i < rs.ActiveSplit {
			dispTime = rs.SplitTimes[i]
			currentTotal += rs.SplitTimes[i]
			if targetTime > 0 {
				segmentColor = splitColor(dispTime, targetTime, rs.ILBestTimes[i])
			}
		} else if i == rs.ActiveSplit && rs.Started {
			dispTime = rs.Elapsed
			currentTotal += rs.Elapsed
			if targetTime > 0 {
				segmentColor = splitColor(dispTime, targetTime, rs.ILBestTimes[i])
			}
		}

		timeStr := "--:--:--.---"
		if (rs.Started && i <= rs.ActiveSplit) || i < rs.ActiveSplit {
			timeStr = formatDuration(dispTime)
		}

		p, pc := pctColor(rs.ILBestTimes[i], rs.FullRunBestTimes[i], Yellow)
		wrp, wrpc := pctColor(rs.WRTimes[i], rs.ILBestTimes[i], Red)

		fmt.Fprintf(&sb,
			"  %-10s Time: %s%-12s%s (Run Best: %s%-12s%s %s%d%%%s IL Best: %s%-12s%s WR: %s%s%s %s%d%%%s)\033[K\n",
			splitName,
			segmentColor, timeStr, Reset,
			Purple, formatOrNone(rs.FullRunBestTimes[i]), Reset, pc, p, Reset,
			Purple, formatOrNone(rs.ILBestTimes[i]), Reset,
			wrpc, formatOrNone(rs.WRTimes[i]), Reset, wrpc, wrp, Reset,
		)
	}

	if rs.ILMode == 0 {
		sb.WriteString(pad("", "-", DisplayWidth, false))
		sb.WriteString("\033[K\n")

		totalColor := Cyan
		if rs.TotalBest > 0 && rs.Started {
			if currentTotal <= rs.TotalBest {
				totalColor = Green
			} else {
				totalColor = Red
			}
		}

		p, pc := pctColor(rs.SOB, rs.TotalBest, Yellow)
		wrp, wrpc := pctColor(rs.WRSob, rs.SOB, Red)

		fmt.Fprintf(&sb,
			"  %-10s Time: %s%-12s%s (Best Total: %s%-12s%s) (SOB: %s%-12s%s %s%d%%%s WR: %s%s%s %s%d%%%s)\033[K\n",
			"TOTAL:",
			totalColor, formatDuration(currentTotal), Reset,
			Purple, formatOrNone(rs.TotalBest), Reset,
			Purple, formatDuration(rs.SOB), Reset, pc, p, Reset,
			wrpc, formatDuration(rs.WRSob), Reset, wrpc, wrp, Reset,
		)
	}

	sb.WriteString(pad("", "=", DisplayWidth, false))
	sb.WriteString("\033[K\n")
	sb.WriteString(statusLine(rs))
	sb.WriteString("\033[J")

	fmt.Print(sb.String())
}

// statusLine returns the footer status string based on current run state.
func statusLine(rs *RunState) string {
	switch {
	case !rs.Started:
		return fmt.Sprintf("Status: %sSTOPPED / READY%s", White, Reset)
	case rs.Paused:
		next := fmt.Sprintf("Split %d", rs.ActiveSplit+2)
		if rs.ILMode > 0 {
			next = "Done"
		}
		return fmt.Sprintf("Status: %sPAUSED (Next WASD starts %s)%s", Yellow, next, Reset)
	default:
		return fmt.Sprintf("Status: %sRUNNING%s", Green, Reset)
	}
}

// splitColor returns a color based on how the current time compares to targets.
func splitColor(actual, runBest, ilBest time.Duration) string {
	if actual > runBest {
		return Red
	}
	if actual > ilBest {
		return Yellow
	}
	return Green
}

// pctColor returns a percentage and its display color given a numerator and denominator.
// normalColor is shown when the percentage is not 100.
func pctColor(num, denom time.Duration, normalColor string) (int, string) {
	if denom == 0 {
		return 0, normalColor
	}
	p := int(float64(num) / float64(denom) * 100.0)
	if p >= 100 {
		return p, Green
	}
	return p, normalColor
}

// formatOrNone returns a formatted duration string or "NONE" if the value is zero.
func formatOrNone(d time.Duration) string {
	if d > 0 {
		return formatDuration(d)
	}
	return "NONE"
}

// formatDuration formats a duration as HH:MM:SS.mmm.
func formatDuration(d time.Duration) string {
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	ms := int(d.Milliseconds()) % 1000
	return fmt.Sprintf("%02d:%02d:%02d.%03d", h, m, s, ms)
}

// pad centers text within a field of totalSize, filling with padChar.
// If newline is true, a newline is appended.
func pad(text, padChar string, totalSize int, newline bool) string {
	if len(text) >= totalSize {
		return text
	}
	totalPadding := totalSize - len(text)
	left := strings.Repeat(padChar, totalPadding/2)
	right := strings.Repeat(padChar, totalPadding-totalPadding/2)
	result := left + text + right
	if newline {
		return result + "\n"
	}
	return result
}
