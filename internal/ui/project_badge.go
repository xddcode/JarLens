package ui

import (
	"path/filepath"
	"strings"
	"unicode"

	vivaui "github.com/vivaui/viva/ui"
)

func projectInitials(path string) string {
	name := strings.TrimSuffix(filepathBase(path), filepath.Ext(path))
	words := strings.FieldsFunc(name, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
	var initials []rune
	for _, word := range words {
		for _, r := range word {
			initials = append(initials, unicode.ToUpper(r))
			break
		}
		if len(initials) == 2 {
			break
		}
	}
	if len(initials) == 0 {
		return "J"
	}
	return string(initials)
}

func projectBadgeColor(path string) vivaui.Color {
	path = filepathBase(path)
	palette := [...]vivaui.Color{
		vivaui.RGB24(0xe08855), vivaui.RGB24(0xb08b14), vivaui.RGB24(0xa1a359),
		vivaui.RGB24(0x3b92b8), vivaui.RGB24(0x3574f0), vivaui.RGB24(0xc84d8f),
		vivaui.RGB24(0x955ae0), vivaui.RGB24(0x24a394), vivaui.RGB24(0x5fad65),
	}
	hash := uint32(2166136261)
	for _, value := range []byte(path) {
		hash = (hash ^ uint32(value)) * 16777619
	}
	return palette[hash%uint32(len(palette))]
}

func paintProjectBadge(canvas *vivaui.Canvas, row vivaui.Rect, path string) {
	if row.Height <= 0 {
		return
	}
	bounds := vivaui.Rect{X: row.X + 6, Y: row.Y + (row.Height-20)/2, Width: 20, Height: 20}
	canvas.Rounded(bounds, 4, projectBadgeColor(path))
	canvas.LabelCenter(projectInitials(path), bounds, vivaui.TextStyle{
		Size: 11, Weight: vivaui.WeightSemibold, Color: vivaui.RGB24(0xffffff),
	})
}
