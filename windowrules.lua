hl.window_rule({
	match = {
		class = "^TIMER$",
	},
	pin = true,
	float = true,
	no_focus = true,
	no_initial_focus = true,
	size = { "940", "180" },
	move = { "0", "0" },
	border_size = 0,
})
hl.window_rule({
	match = { class = "^TIMER$", tag = "HIDE" },
	opacity = "0 override",
})
hl.window_rule({
	match = { class = "^TIMER$" },
	opacity = "1 override",
})
