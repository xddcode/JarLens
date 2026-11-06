package ui

import vivaui "github.com/vivaui/viva/ui"

// Package mark: JetBrains IntelliJ Platform, Apache-2.0.
var searchGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(10.5, 4.5),
		vivaui.CubicTo(7.2, 4.5, 4.5, 7.2, 4.5, 10.5),
		vivaui.CubicTo(4.5, 13.8, 7.2, 16.5, 10.5, 16.5),
		vivaui.CubicTo(13.8, 16.5, 16.5, 13.8, 16.5, 10.5),
		vivaui.CubicTo(16.5, 7.2, 13.8, 4.5, 10.5, 4.5),
		vivaui.MoveTo(16, 16), vivaui.LineTo(22, 22),
	},
}

var infoGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 3),
		vivaui.CubicTo(7, 3, 3, 7, 3, 12),
		vivaui.CubicTo(3, 17, 7, 21, 12, 21),
		vivaui.CubicTo(17, 21, 21, 17, 21, 12),
		vivaui.CubicTo(21, 7, 17, 3, 12, 3), vivaui.ClosePath(),
		vivaui.MoveTo(12, 11), vivaui.LineTo(12, 17),
		vivaui.MoveTo(12, 7), vivaui.LineTo(12, 7.1),
	},
}

var homeGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(3, 11), vivaui.LineTo(12, 4), vivaui.LineTo(21, 11),
		vivaui.MoveTo(5, 10), vivaui.LineTo(5, 20), vivaui.LineTo(19, 20), vivaui.LineTo(19, 10),
		vivaui.MoveTo(9, 20), vivaui.LineTo(9, 14), vivaui.LineTo(15, 14), vivaui.LineTo(15, 20),
	},
}

var openGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(3, 18), vivaui.LineTo(3, 6), vivaui.LineTo(9, 6),
		vivaui.LineTo(11, 8), vivaui.LineTo(20, 8), vivaui.LineTo(20, 11),
		vivaui.MoveTo(3, 18), vivaui.LineTo(6, 11), vivaui.LineTo(22, 11),
		vivaui.LineTo(19, 18), vivaui.ClosePath(),
	},
}

var saveGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(5, 3), vivaui.LineTo(17, 3), vivaui.LineTo(21, 7),
		vivaui.LineTo(21, 21), vivaui.LineTo(3, 21), vivaui.LineTo(3, 3), vivaui.ClosePath(),
		vivaui.MoveTo(7, 3), vivaui.LineTo(7, 10), vivaui.LineTo(17, 10), vivaui.LineTo(17, 3),
		vivaui.MoveTo(8, 21), vivaui.LineTo(8, 15), vivaui.LineTo(16, 15), vivaui.LineTo(16, 21),
	},
}

var cancelGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(7, 7), vivaui.LineTo(17, 17),
		vivaui.MoveTo(17, 7), vivaui.LineTo(7, 17),
	},
}

var arrowUpGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 20), vivaui.LineTo(12, 4),
		vivaui.MoveTo(5, 11), vivaui.LineTo(12, 4), vivaui.LineTo(19, 11),
	},
}

var arrowDownGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 4), vivaui.LineTo(12, 20),
		vivaui.MoveTo(5, 13), vivaui.LineTo(12, 20), vivaui.LineTo(19, 13),
	},
}

var moreVerticalGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 3, Cap: vivaui.CapRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 5), vivaui.LineTo(12, 5.1),
		vivaui.MoveTo(12, 12), vivaui.LineTo(12, 12.1),
		vivaui.MoveTo(12, 19), vivaui.LineTo(12, 19.1),
	},
}

var menuGlyph = vivaui.Icon{
	ViewBox: 24, Stroke: 2, Cap: vivaui.CapRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(4, 6), vivaui.LineTo(20, 6),
		vivaui.MoveTo(4, 12), vivaui.LineTo(20, 12),
		vivaui.MoveTo(4, 18), vivaui.LineTo(20, 18),
	},
}

var splitRightGlyph = vivaui.Icon{
	ViewBox: 16, Stroke: 1.2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(3, 2), vivaui.LineTo(13, 2), vivaui.LineTo(13, 14), vivaui.LineTo(3, 14), vivaui.ClosePath(),
		vivaui.MoveTo(8, 2), vivaui.LineTo(8, 14),
	},
}

var splitDownGlyph = vivaui.Icon{
	ViewBox: 16, Stroke: 1.2, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(2, 3), vivaui.LineTo(14, 3), vivaui.LineTo(14, 13), vivaui.LineTo(2, 13), vivaui.ClosePath(),
		vivaui.MoveTo(2, 8), vivaui.LineTo(14, 8),
	},
}

var xmlGlyph = vivaui.Icon{
	ViewBox: 16, Stroke: 1.3, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(5, 4), vivaui.LineTo(1, 8), vivaui.LineTo(5, 12),
		vivaui.MoveTo(11, 4), vivaui.LineTo(15, 8), vivaui.LineTo(11, 12),
		vivaui.MoveTo(9.5, 2), vivaui.LineTo(6.5, 14),
	},
}

var chevronRightGlyph = vivaui.Icon{
	ViewBox: 16, Stroke: 1.5, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{vivaui.MoveTo(6, 4.5), vivaui.LineTo(9.5, 8), vivaui.LineTo(6, 11.5)},
}

var chevronDownGlyph = vivaui.Icon{
	ViewBox: 16, Stroke: 1.5, Cap: vivaui.CapRound, Join: vivaui.JoinRound,
	Commands: []vivaui.PathCommand{vivaui.MoveTo(4.5, 6.25), vivaui.LineTo(8, 9.75), vivaui.LineTo(11.5, 6.25)},
}

var sunGlyph = vivaui.Icon{
	ViewBox: 24,
	Stroke:  2,
	Cap:     vivaui.CapRound,
	Join:    vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 8),
		vivaui.CubicTo(9.8, 8, 8, 9.8, 8, 12),
		vivaui.CubicTo(8, 14.2, 9.8, 16, 12, 16),
		vivaui.CubicTo(14.2, 16, 16, 14.2, 16, 12),
		vivaui.CubicTo(16, 9.8, 14.2, 8, 12, 8), vivaui.ClosePath(),
		vivaui.MoveTo(12, 2), vivaui.LineTo(12, 4),
		vivaui.MoveTo(12, 20), vivaui.LineTo(12, 22),
		vivaui.MoveTo(2, 12), vivaui.LineTo(4, 12),
		vivaui.MoveTo(20, 12), vivaui.LineTo(22, 12),
		vivaui.MoveTo(4.9, 4.9), vivaui.LineTo(6.3, 6.3),
		vivaui.MoveTo(17.7, 17.7), vivaui.LineTo(19.1, 19.1),
		vivaui.MoveTo(4.9, 19.1), vivaui.LineTo(6.3, 17.7),
		vivaui.MoveTo(17.7, 6.3), vivaui.LineTo(19.1, 4.9),
	},
}

var moonGlyph = vivaui.Icon{
	ViewBox: 24,
	Stroke:  2,
	Cap:     vivaui.CapRound,
	Join:    vivaui.JoinRound,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(20, 15.5),
		vivaui.CubicTo(17, 17.5, 13.5, 16, 11.5, 12.5),
		vivaui.CubicTo(10, 9.5, 11, 5.5, 14, 3.5),
		vivaui.CubicTo(8, 3.5, 4, 7.5, 4, 13),
		vivaui.CubicTo(4, 18, 8, 21, 13, 21),
		vivaui.CubicTo(16.5, 21, 19, 18.5, 20, 15.5), vivaui.ClosePath(),
	},
}

// GitHub mark: Simple Icons, CC0-1.0. Brand rights remain with their owner.
var githubGlyph = vivaui.Icon{
	ViewBox: 24,
	Commands: []vivaui.PathCommand{
		vivaui.MoveTo(12, 0.297),
		vivaui.CubicTo(5.37, 0.297, 0, 5.67, 0, 12.297),
		vivaui.CubicTo(0, 17.6, 3.438, 22.097, 8.205, 23.682),
		vivaui.CubicTo(8.805, 23.795, 9.025, 23.424, 9.025, 23.105),
		vivaui.CubicTo(9.025, 22.82, 9.015, 22.065, 9.01, 21.065),
		vivaui.CubicTo(5.672, 21.789, 4.968, 19.455, 4.968, 19.455),
		vivaui.CubicTo(4.422, 18.07, 3.633, 17.7, 3.633, 17.7),
		vivaui.CubicTo(2.546, 16.956, 3.717, 16.971, 3.717, 16.971),
		vivaui.CubicTo(4.922, 17.055, 5.555, 18.207, 5.555, 18.207),
		vivaui.CubicTo(6.625, 20.042, 8.364, 19.512, 9.05, 19.205),
		vivaui.CubicTo(9.158, 18.429, 9.467, 17.9, 9.81, 17.6),
		vivaui.CubicTo(7.145, 17.3, 4.344, 16.268, 4.344, 11.67),
		vivaui.CubicTo(4.344, 10.36, 4.809, 9.29, 5.579, 8.45),
		vivaui.CubicTo(5.444, 8.147, 5.039, 6.927, 5.684, 5.274),
		vivaui.CubicTo(5.684, 5.274, 6.689, 4.952, 8.984, 6.504),
		vivaui.CubicTo(9.944, 6.237, 10.964, 6.105, 11.984, 6.099),
		vivaui.CubicTo(13.004, 6.105, 14.024, 6.237, 14.984, 6.504),
		vivaui.CubicTo(17.264, 4.952, 18.269, 5.274, 18.269, 5.274),
		vivaui.CubicTo(18.914, 6.927, 18.509, 8.147, 18.389, 8.45),
		vivaui.CubicTo(19.154, 9.29, 19.619, 10.36, 19.619, 11.67),
		vivaui.CubicTo(19.619, 16.28, 16.814, 17.295, 14.144, 17.59),
		vivaui.CubicTo(14.564, 17.95, 14.954, 18.686, 14.954, 19.81),
		vivaui.CubicTo(14.954, 21.416, 14.939, 22.706, 14.939, 23.096),
		vivaui.CubicTo(14.939, 23.411, 15.149, 23.786, 15.764, 23.666),
		vivaui.CubicTo(20.565, 22.092, 24, 17.592, 24, 12.297),
		vivaui.CubicTo(24, 5.67, 18.627, 0.297, 12, 0.297),
	},
}
