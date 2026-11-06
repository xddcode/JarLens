package ui

import (
	"strings"
	"testing"
	"time"

	vivaui "github.com/vivaui/viva/ui"
)

func TestWorkspaceTooltipDelayAndStyle(t *testing.T) {
	now := time.Duration(0)
	var hover workspaceTooltip
	paint := func() (*vivaui.Canvas, bool) {
		var canvas vivaui.Canvas
		canvas.SetClock(func() time.Duration { return now })
		canvas.BeginFrame()
		canvas.HitWithOptions(vivaui.Item(7), vivaui.Rect{X: 40, Y: 40, Width: 100, Height: 28}, "文档", vivaui.HitOptions{Hint: "fixture.jar!/example/Hello.class"})
		canvas.SetInteraction(vivaui.Item(7), vivaui.ID{}, vivaui.ID{}, false)
		painted := hover.Paint(&canvas, AppearanceDark, vivaui.Rect{Width: 800, Height: 600}, workspaceTooltipDelay)
		canvas.EndFrame()
		return &canvas, painted
	}
	if _, painted := paint(); painted {
		t.Fatal("workspace tooltip appeared without its hover delay")
	}
	now = workspaceTooltipDelay - time.Millisecond
	if _, painted := paint(); painted {
		t.Fatal("workspace tooltip appeared before 600ms")
	}
	now = workspaceTooltipDelay
	canvas, painted := paint()
	if !painted {
		t.Fatal("workspace tooltip did not appear after 600ms")
	}
	for _, text := range canvas.Texts() {
		if text.Value == "fixture.jar!/example/Hello.class" {
			if text.Style.Size != 13 || text.Style.Color != vivaui.RGB24(0xd1d3d9) || text.Bounds.Y < 74 {
				t.Fatalf("workspace tooltip text does not match the Islands style: %+v", text)
			}
			found := false
			for _, rectangle := range canvas.Rectangles() {
				if rectangle.Color == vivaui.RGB24(0x33353b) && rectangle.Radius == 6 {
					found = true
				}
			}
			if !found {
				t.Fatal("workspace tooltip does not have its rounded dark floating surface")
			}
			return
		}
	}
	t.Fatal("workspace tooltip text was not painted")
}

func TestWorkspaceTooltipWrapsLongPaths(t *testing.T) {
	var canvas vivaui.Canvas
	canvas.BeginFrame()
	content := paintWorkspaceTooltip(&canvas, AppearanceDark, vivaui.Rect{Width: 800, Height: 600}, vivaui.Rect{X: 40, Y: 40, Width: 100, Height: 28}, strings.Repeat("long-path/", 100))
	canvas.EndFrame()
	if content.Width > 720 || len(canvas.Texts()) < 2 {
		t.Fatalf("long tooltip path was not wrapped inside its 720px surface: %+v, %d lines", content, len(canvas.Texts()))
	}
}
