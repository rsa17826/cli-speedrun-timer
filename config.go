package main

// ANSI color escape codes.
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Cyan   = "\033[36m"
	Purple = "\033[35m"
	White  = "\033[37m"
)

// Machine-wide layout constants (not game specific).
const (
	DisplayWidth = 104
)

// Record file names, stored inside each game's DataDir.
const (
	FullRunFile = "times_full_run"
	ILFile      = "times_il_all"
	WRFile      = "wrs"
)
