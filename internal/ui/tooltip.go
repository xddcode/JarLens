package ui

import (
	"time"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

const (
	workspaceTooltipDelay = 600 * time.Millisecond
	workspaceTooltipWarm  = 400 * time.Millisecond
)

type workspaceTooltip struct {
	target     vivaui.ID
	shown      vivaui.ID
	suppressed vivaui.ID
	since      time.Duration
	hiddenAt   time.Duration
	focus      bool
}

func (hover *workspaceTooltip) ShowOnFocus() { hover.focus = true }

func (hover *workspaceTooltip) Dismiss() bool {
	shown := !hover.shown.IsZero()
	if shown {
		hover.suppressed = hover.shown
	}
	hover.target, hover.shown = vivaui.ID{}, vivaui.ID{}
	return shown
}

func (hover *workspaceTooltip) Paint(canvas *vivaui.Canvas, appearance Appearance, viewport vivaui.Rect, delay time.Duration) bool {
	now := canvas.Now()
	hovered := canvas.Hovered()
	if !canvas.Pressed().IsZero() {
		if !hover.shown.IsZero() {
			hover.suppressed = hover.shown
		}
		hover.target, hover.shown = vivaui.ID{}, vivaui.ID{}
		hover.hiddenAt = now
		return false
	}
	if !hover.suppressed.IsZero() {
		if hovered == hover.suppressed {
			hover.target, hover.shown = vivaui.ID{}, vivaui.ID{}
			return false
		}
		hover.suppressed = vivaui.ID{}
	}
	id := hovered
	immediate := false
	if hover.focus {
		if focused := canvas.FocusVisible(); !focused.IsZero() {
			id, immediate = focused, true
		}
	}
	if id.IsZero() {
		if !hover.shown.IsZero() {
			hover.hiddenAt = now
		}
		hover.target, hover.shown = vivaui.ID{}, vivaui.ID{}
		return false
	}

	var hit vivaui.Hit
	found := false
	hits := canvas.Hits()
	for index := len(hits) - 1; index >= 0; index-- {
		if hits[index].ID == id && !hits[index].Block {
			hit, found = hits[index], hits[index].Hint != ""
			break
		}
	}
	if !found {
		hover.target, hover.shown = vivaui.ID{}, vivaui.ID{}
		return false
	}
	if hover.target != id {
		immediate = immediate || !hover.shown.IsZero() || hover.hiddenAt > 0 && now-hover.hiddenAt <= workspaceTooltipWarm
		hover.target, hover.since = id, now
	}
	if immediate {
		delay = 0
	}
	if remaining := delay - (now - hover.since); remaining > 0 {
		canvas.RequestFrameIn(remaining)
		hover.shown = vivaui.ID{}
		return false
	}
	paintWorkspaceTooltip(canvas, appearance, viewport, hit.Bounds, hit.Hint)
	hover.shown = id
	hover.hiddenAt = 0
	return true
}

func paintWorkspaceTooltip(canvas *vivaui.Canvas, appearance Appearance, viewport, anchor vivaui.Rect, label string) vivaui.Rect {
	if label == "" || viewport.Width <= 32 || viewport.Height <= 0 {
		return vivaui.Rect{}
	}
	background := vivaui.RGB24(0x33353b)
	border := background
	foreground := vivaui.RGB24(0xd1d3d9)
	if appearance == AppearanceLight {
		background = vivaui.RGB24(0xffffff)
		border = vivaui.RGB24(0xd1d3d9)
		foreground = vivaui.RGB24(0x000000)
	}
	const (
		fontSize   = float32(13)
		lineHeight = float32(16)
		padX       = float32(10)
		padY       = float32(6)
		borderSize = float32(1)
	)
	maxWidth := min(float32(720), viewport.Width-32)
	lines := wrapWorkspaceTooltip(canvas, label, maxWidth-2*(padX+borderSize), fontSize)
	if len(lines) == 0 {
		return vivaui.Rect{}
	}
	textWidth := maxWidth - 2*(padX+borderSize)
	if len(lines) == 1 {
		textWidth = canvas.TextWidth(lines[0], fontSize)
	}
	size := vivaui.Size{
		Width:  textWidth + 2*(padX+borderSize),
		Height: float32(len(lines))*lineHeight + 2*(padY+borderSize),
	}
	content, _ := components.PlaceOverlay(anchor, size, viewport, components.OverlayPlacement{
		Side: components.OverlayBottom, Align: components.OverlayAlignCenter, SideOffset: 6, Margin: 6,
	})
	canvas.PushLayer(4)
	canvas.Elevate(content, 6, vivaui.ShadowMD)
	canvas.RoundedBorder(content, 6, borderSize, border, background)
	for index, line := range lines {
		canvas.Label(line, vivaui.Rect{
			X: content.X + padX + borderSize, Y: content.Y + padY + borderSize + float32(index)*lineHeight,
			Width: content.Width - 2*(padX+borderSize), Height: lineHeight,
		}, vivaui.TextStyle{Size: fontSize, Color: foreground})
	}
	canvas.PopLayer()
	return content
}

func wrapWorkspaceTooltip(canvas *vivaui.Canvas, value string, maxWidth, fontSize float32) []string {
	if canvas.TextWidth(value, fontSize) <= maxWidth {
		return []string{value}
	}
	runes := []rune(value)
	var lines []string
	for start := 0; start < len(runes); {
		end := start + 1
		for end < len(runes) && runes[end] != '\n' && canvas.TextWidth(string(runes[start:end+1]), fontSize) <= maxWidth {
			end++
		}
		lines = append(lines, string(runes[start:end]))
		start = end
		if start < len(runes) && runes[start] == '\n' {
			start++
		}
	}
	return lines
}
