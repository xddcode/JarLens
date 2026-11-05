package ui

import vivaui "github.com/vivaui/viva/ui"

type Appearance uint8

const (
	AppearanceDark Appearance = iota
	AppearanceLight
)

func WorkspaceTheme(appearance Appearance) vivaui.Theme {
	if appearance == AppearanceLight {
		return lightTheme()
	}
	return darkTheme()
}

func darkTheme() vivaui.Theme {
	theme := vivaui.DarkTheme()
	theme.Background = vivaui.RGB24(0x191a1c)
	theme.Foreground = vivaui.RGB24(0xd1d3d9)
	theme.Card = vivaui.RGB24(0x26282c)
	theme.CardForeground = theme.Foreground
	theme.CardElevated = vivaui.RGB24(0x212326)
	theme.Popover = vivaui.RGB24(0x26282c)
	theme.PopoverForeground = theme.Foreground
	theme.Muted = vivaui.RGB24(0x2e2f30)
	theme.MutedForeground = vivaui.RGB24(0x9fa2a8)
	theme.Border = vivaui.RGB24(0x40434a)
	theme.Hover = vivaui.RGB24(0x2e2f30)
	theme.Input = vivaui.RGB24(0x191a1c)
	theme.Primary = vivaui.RGB24(0x3871e1)
	theme.PrimaryForeground = vivaui.RGB24(0xffffff)
	theme.Selection = vivaui.RGB24(0x2a4371)
	theme.SelectionForeground = vivaui.RGB24(0xdfe1e5)
	theme.Secondary = vivaui.RGB24(0x26282c)
	theme.SecondaryForeground = theme.Foreground
	theme.Accent = vivaui.RGB24(0x2a4371)
	theme.AccentForeground = vivaui.RGB24(0xdfe1e5)
	theme.Ring = vivaui.RGB24(0x3871e1)
	theme.Success = vivaui.RGB24(0x6db083)
	theme.Danger = vivaui.RGB24(0xf57e84)
	theme.DangerForeground = vivaui.RGB24(0x191a1c)
	setWorkspaceMetrics(&theme)
	return theme
}

func lightTheme() vivaui.Theme {
	theme := vivaui.LightTheme()
	theme.Background = vivaui.RGB24(0xffffff)
	theme.Foreground = vivaui.RGB24(0x000000)
	theme.Card = vivaui.RGB24(0xe9eaee)
	theme.CardForeground = theme.Foreground
	theme.CardElevated = vivaui.RGB24(0xf7f8f9)
	theme.Popover = vivaui.RGB24(0xffffff)
	theme.PopoverForeground = theme.Foreground
	theme.Muted = vivaui.RGB24(0xededed)
	theme.MutedForeground = vivaui.RGB24(0x5f6269)
	theme.Border = vivaui.RGB24(0xd1d3d9)
	theme.Hover = vivaui.RGB24(0xededed)
	theme.Input = vivaui.RGB24(0xffffff)
	theme.Primary = vivaui.RGB24(0x3871e1)
	theme.PrimaryForeground = vivaui.RGB24(0xffffff)
	theme.Selection = vivaui.RGB24(0xd0dffe)
	theme.SelectionForeground = vivaui.RGB24(0x000000)
	theme.Secondary = vivaui.RGB24(0xf7f8f9)
	theme.SecondaryForeground = theme.Foreground
	theme.Accent = vivaui.RGB24(0xd0dffe)
	theme.AccentForeground = vivaui.RGB24(0x000000)
	theme.Ring = vivaui.RGB24(0x3871e1)
	theme.Success = vivaui.RGB24(0x338555)
	theme.Danger = vivaui.RGB24(0xc54e58)
	theme.DangerForeground = vivaui.RGB24(0xffffff)
	setWorkspaceMetrics(&theme)
	return theme
}

func setWorkspaceMetrics(theme *vivaui.Theme) {
	theme.Metrics.ControlHeight = 28
	theme.Metrics.RadiusControl = 4
	theme.Metrics.RadiusSurface = 10
	theme.Metrics.FontSizeS = 12
	theme.Metrics.FontSizeM = 13
	theme.Metrics.FontSizeL = 15
	theme.Metrics.PadControlX = 10
	theme.Metrics.PadSurface = 12
	theme.Metrics.GapItem = 8
}
