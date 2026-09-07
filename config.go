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

// Layout / game constants.
const (
	TotalSplits  = 5
	ScreenWidth  = 1920
	ScreenHeight = 1080
	DisplayWidth = 104
)

// File and directory paths.
const (
	FullRunPath = "./times_full_run"
	ILPath      = "./times_il_all"
	WRPath      = "./wrs"
	WatchDir    = "/home/nyix/projects/mathbreakers-ghost-speedrun-mod/mathbreakers"
	TargetFile  = "level_cleared.txt"
	PIDFile     = "/tmp/mathbreakers_pid_a"
)

// LevelPositions maps a level index to its [x, y] click coordinates on screen.
var LevelPositions = [][2]int32{
	{546, 293},
	{549, 555},
	{944, 556},
	{547, 808},
	{935, 814},
}
