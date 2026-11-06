package ui

import (
	"path/filepath"
	"strings"

	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
)

func paintFileIcon(f *flex.Frame, filePath string, kind archive.NodeKind, expanded bool, size float32) flex.Box {
	if kind == archive.NodePackage {
		glyph := materialFolderGlyph
		if expanded {
			glyph = materialFolderOpenGlyph
		}
		return f.Icon(glyph).Size(size, size)
	}
	if kind == archive.NodeClass {
		return materialTextBadge(f, "CL", vivaui.RGB24(0x8d6e63), vivaui.RGB24(0xffffff), 7.4, size)
	}
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".java":
		return materialTextBadge(f, "J", vivaui.RGB24(0xe53935), vivaui.RGB24(0xffffff), 9.5, size)
	case ".json":
		return materialTextBadge(f, "{}", vivaui.RGB24(0xfbc02d), vivaui.RGB24(0x5d4037), 9, size)
	case ".yaml", ".yml":
		return materialTextBadge(f, "YML", vivaui.RGB24(0xe53935), vivaui.RGB24(0xffffff), 5.6, size)
	case ".properties", ".ini", ".conf", ".cfg":
		return materialTextBadge(f, "INI", vivaui.RGB24(0x78909c), vivaui.RGB24(0xffffff), 5.6, size)
	case ".xml", ".xsd", ".html", ".htm":
		box := f.Row().Size(size, size).Radius(size * 3 / 16).Background(vivaui.RGB24(0xff9800)).Center()
		box.Children(func() {
			f.Icon(xmlGlyph).Size(size*3/4, size*3/4).Color(vivaui.RGB24(0xffffff))
		})
		return box
	default:
		return f.Icon(materialTextGlyph).Size(size, size)
	}
}

func materialTextBadge(f *flex.Frame, label string, background, foreground vivaui.Color, fontSize, size float32) flex.Box {
	box := f.Row().Size(size, size).Radius(size * 3 / 16).Background(background).Center()
	box.Children(func() {
		f.Text(label).FontSize(fontSize * size / 16).Bold().Color(foreground)
	})
	return box
}

var materialFolderGlyph = vivaui.Icon{
	ViewBox: 16,
	Layers: []vivaui.IconLayer{{Fill: vivaui.RGB24(0x90a4ae), Commands: []vivaui.PathCommand{
		vivaui.MoveTo(1.5, 2.5), vivaui.LineTo(6.1, 2.5), vivaui.LineTo(7.6, 4), vivaui.LineTo(14.5, 4),
		vivaui.QuadTo(15.5, 4, 15.5, 5), vivaui.LineTo(15.5, 12.5),
		vivaui.QuadTo(15.5, 13.5, 14.5, 13.5), vivaui.LineTo(1.5, 13.5),
		vivaui.QuadTo(0.5, 13.5, 0.5, 12.5), vivaui.LineTo(0.5, 3.5),
		vivaui.QuadTo(0.5, 2.5, 1.5, 2.5), vivaui.ClosePath(),
	}}},
}

var materialFolderOpenGlyph = vivaui.Icon{
	ViewBox: 16,
	Layers: []vivaui.IconLayer{
		{Fill: vivaui.RGB24(0x90a4ae), Commands: []vivaui.PathCommand{
			vivaui.MoveTo(1.5, 2.5), vivaui.LineTo(6.1, 2.5), vivaui.LineTo(7.6, 4), vivaui.LineTo(14, 4),
			vivaui.QuadTo(15, 4, 15, 5), vivaui.LineTo(15, 6.5), vivaui.LineTo(4.2, 6.5),
			vivaui.LineTo(2, 13.5), vivaui.LineTo(1.5, 13.5), vivaui.QuadTo(0.5, 13.5, 0.5, 12.5),
			vivaui.LineTo(0.5, 3.5), vivaui.QuadTo(0.5, 2.5, 1.5, 2.5), vivaui.ClosePath(),
		}},
		{Fill: vivaui.RGBA(144, 164, 174, 0.75), Commands: []vivaui.PathCommand{
			vivaui.MoveTo(4.2, 7), vivaui.LineTo(15.5, 7), vivaui.LineTo(13.3, 13.5), vivaui.LineTo(2, 13.5), vivaui.ClosePath(),
		}},
	},
}

var materialTextGlyph = vivaui.Icon{
	ViewBox: 16,
	Layers: []vivaui.IconLayer{
		{Fill: vivaui.RGB24(0x90a4ae), Commands: []vivaui.PathCommand{
			vivaui.MoveTo(4, 1), vivaui.LineTo(9.6, 1), vivaui.LineTo(13, 4.4), vivaui.LineTo(13, 14),
			vivaui.QuadTo(13, 15, 12, 15), vivaui.LineTo(4, 15), vivaui.QuadTo(3, 15, 3, 14),
			vivaui.LineTo(3, 2), vivaui.QuadTo(3, 1, 4, 1), vivaui.ClosePath(),
		}},
		{Fill: vivaui.RGBA(255, 255, 255, 0.45), Commands: []vivaui.PathCommand{
			vivaui.MoveTo(9.6, 1), vivaui.LineTo(9.6, 4.4), vivaui.LineTo(13, 4.4), vivaui.ClosePath(),
		}},
		{Fill: vivaui.RGB24(0xffffff), Commands: []vivaui.PathCommand{
			vivaui.MoveTo(5, 7), vivaui.LineTo(11, 7), vivaui.LineTo(11, 8), vivaui.LineTo(5, 8), vivaui.ClosePath(),
			vivaui.MoveTo(5, 9.5), vivaui.LineTo(11, 9.5), vivaui.LineTo(11, 10.5), vivaui.LineTo(5, 10.5), vivaui.ClosePath(),
			vivaui.MoveTo(5, 12), vivaui.LineTo(9, 12), vivaui.LineTo(9, 13), vivaui.LineTo(5, 13), vivaui.ClosePath(),
		}},
	},
}
