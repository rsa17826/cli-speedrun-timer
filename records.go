package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"time"
)

// loadBestTimes reads all three record files (full-run, IL, WR) into rs.
func loadBestTimes(rs *RunState) {
	loadFullRunTimes(rs)
	loadILTimes(rs)
	loadWRTimes(rs)
}

func loadFullRunTimes(rs *RunState) {
	file, err := os.Open(FullRunPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for i := 0; i < TotalSplits && scanner.Scan(); i++ {
		if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
			rs.FullRunBestTimes[i] = time.Duration(ms) * time.Millisecond
		}
	}
	// The line after the splits stores the total best run time.
	if scanner.Scan() {
		if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
			rs.TotalBest = time.Duration(ms) * time.Millisecond
		}
	}
}

func loadILTimes(rs *RunState) {
	file, err := os.Open(ILPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rs.SOB = 0
	for i := 0; i < TotalSplits && scanner.Scan(); i++ {
		if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
			rs.ILBestTimes[i] = time.Duration(ms) * time.Millisecond
			rs.SOB += rs.ILBestTimes[i]
		}
	}
}

func loadWRTimes(rs *RunState) {
	file, err := os.Open(WRPath)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	rs.WRSob = 0
	for i := 0; i < TotalSplits && scanner.Scan(); i++ {
		if ms, err := strconv.ParseInt(scanner.Text(), 10, 64); err == nil {
			rs.WRTimes[i] = time.Duration(ms) * time.Millisecond
			rs.WRSob += rs.WRTimes[i]
		}
	}
}

// saveFile persists the relevant best times for the given target ("il" or "full").
func saveFile(rs *RunState, target string) {
	switch target {
	case "il":
		saveILTimes(rs)
	case "full":
		saveFullRunTimes(rs)
	}
}

func saveILTimes(rs *RunState) {
	file, err := os.Create(ILPath)
	if err != nil {
		return
	}
	defer file.Close()
	for _, t := range rs.ILBestTimes {
		fmt.Fprintf(file, "%d\n", t.Milliseconds())
	}
}

func saveFullRunTimes(rs *RunState) {
	file, err := os.Create(FullRunPath)
	if err != nil {
		return
	}
	defer file.Close()
	for _, t := range rs.FullRunBestTimes {
		fmt.Fprintf(file, "%d\n", t.Milliseconds())
	}
	fmt.Fprintf(file, "%d\n", rs.TotalBest.Milliseconds())
}
