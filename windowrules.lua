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
local timer_visible = hl.window_rule({
	name = "timer-visible",
	match = {
		class = "^TIMER$",
	},
	opacity = "1 override 1 override 1 override",
})

local timer_hidden = hl.window_rule({
	name = "timer-hidden",
	match = {
		class = "^TIMER$",
	},
	opacity = "0 override 0 override 0 override",
})

-- Start hidden.
timer_visible:set_enabled(false)
timer_hidden:set_enabled(true)

hl.on("window.active", function(w)
	if w.class == "explorer.exe" then
		timer_hidden:set_enabled(false)
		timer_visible:set_enabled(true)
	else
		timer_visible:set_enabled(false)
		timer_hidden:set_enabled(true)
	end
end)
