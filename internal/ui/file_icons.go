package ui

import (
	"path/filepath"
	"strings"

	"github.com/vivaui/viva/flex"
	vivaui "github.com/vivaui/viva/ui"
	"github.com/xddcode/jarlens/internal/archive"
	"github.com/xddcode/jarlens/internal/ui/openideaicons"
)

type fileIconID uint8

const (
	fileIconPackage fileIconID = iota
	fileIconClass
	fileIconJava
	fileIconJSON
	fileIconYAML
	fileIconConfig
	fileIconProperties
	fileIconXML
	fileIconHTML
	fileIconXSD
	fileIconText
)

type themedFileIcon struct {
	light vivaui.Icon
	dark  vivaui.Icon
}

var openIDEAFileIcons = [...]themedFileIcon{
	fileIconPackage:    {light: openideaicons.Package, dark: openideaicons.Package_dark},
	fileIconClass:      {light: openIDEAClassIcon(openideaicons.Class, vivaui.RGB24(0x3574f0), vivaui.RGB24(0xe7effd)), dark: openIDEAClassIcon(openideaicons.Class_dark, vivaui.RGB24(0x548af7), vivaui.RGB24(0x25324d))},
	fileIconJava:       {light: openIDEAJavaIcon(openideaicons.Java), dark: openIDEAJavaIcon(openideaicons.Java_dark)},
	fileIconJSON:       {light: openideaicons.Json, dark: openideaicons.Json_dark},
	fileIconYAML:       {light: openideaicons.Yaml, dark: openideaicons.Yaml_dark},
	fileIconConfig:     {light: openideaicons.Config, dark: openideaicons.Config_dark},
	fileIconProperties: {light: openideaicons.Properties, dark: openideaicons.Properties_dark},
	fileIconXML:        {light: openideaicons.Xml, dark: openideaicons.Xml_dark},
	fileIconHTML:       {light: openideaicons.Html, dark: openideaicons.Html_dark},
	fileIconXSD:        {light: openideaicons.Xsd, dark: openideaicons.Xsd_dark},
	fileIconText:       {light: openideaicons.Text, dark: openideaicons.Text_dark},
}

func paintFileIcon(f *flex.Frame, filePath string, kind archive.NodeKind, dark bool, size float32) flex.Box {
	id := fileIconIDFor(filePath, kind)
	icon := f.Icon(openIDEAFileIcon(id, dark)).Size(size, size)
	if id == fileIconJava {
		color := vivaui.RGB24(0xe66d17)
		if dark {
			color = vivaui.RGB24(0xc77d55)
		}
		return icon.Color(color)
	}
	return icon
}

func fileIconIDFor(filePath string, kind archive.NodeKind) fileIconID {
	if kind == archive.NodePackage {
		return fileIconPackage
	}
	if kind == archive.NodeClass {
		return fileIconClass
	}
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".java":
		return fileIconJava
	case ".json":
		return fileIconJSON
	case ".yaml", ".yml":
		return fileIconYAML
	case ".ini", ".conf", ".cfg":
		return fileIconConfig
	case ".properties":
		return fileIconProperties
	case ".xml":
		return fileIconXML
	case ".html", ".htm":
		return fileIconHTML
	case ".xsd":
		return fileIconXSD
	default:
		return fileIconText
	}
}

func openIDEAFileIcon(id fileIconID, dark bool) vivaui.Icon {
	if dark {
		return openIDEAFileIcons[id].dark
	}
	return openIDEAFileIcons[id].light
}

func openIDEAClassIcon(source vivaui.Icon, outline, background vivaui.Color) vivaui.Icon {
	source.Layers = []vivaui.IconLayer{
		{Fill: outline, Commands: openIDEACircle(7)},
		{Fill: background, Commands: openIDEACircle(6)},
		source.Layers[1],
	}
	return source
}

func openIDEACircle(radius float32) []vivaui.PathCommand {
	const kappa = 0.5522848
	offset := radius * kappa
	return []vivaui.PathCommand{
		vivaui.MoveTo(8+radius, 8),
		vivaui.CubicTo(8+radius, 8+offset, 8+offset, 8+radius, 8, 8+radius),
		vivaui.CubicTo(8-offset, 8+radius, 8-radius, 8+offset, 8-radius, 8),
		vivaui.CubicTo(8-radius, 8-offset, 8-offset, 8-radius, 8, 8-radius),
		vivaui.CubicTo(8+offset, 8-radius, 8+radius, 8-offset, 8+radius, 8),
		vivaui.ClosePath(),
	}
}

func openIDEAJavaIcon(source vivaui.Icon) vivaui.Icon {
	commands := make([]vivaui.PathCommand, 0, len(source.Commands)+len(source.Layers[1].Commands))
	commands = append(commands, source.Commands...)
	commands = append(commands, source.Layers[1].Commands...)
	source.Commands = commands
	source.Stroke = 1
	return source
}
