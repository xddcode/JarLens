package ui

import (
	"path/filepath"
	"strings"

	"github.com/vivaui/viva/components"
	vivaui "github.com/vivaui/viva/ui"
)

func languageForPath(filePath string) string {
	name := strings.ToLower(filepath.Base(filePath))
	switch filepath.Ext(name) {
	case ".java", ".class":
		return "java"
	case ".xml", ".xsd", ".html", ".htm":
		return "xml"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".md", ".markdown":
		return "markdown"
	case ".properties", ".ini", ".toml", ".conf", ".cfg", ".mf":
		return "properties"
	default:
		return "text"
	}
}

func highlightDocumentLine(appearance Appearance, language, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
	switch language {
	case "java":
		return highlightJavaLine(appearance, language, line, dst)
	case "xml":
		return highlightXMLLine(appearance, line, dst)
	default:
		return dst
	}
}

func highlightXMLLine(appearance Appearance, line string, dst []components.CodeEditorRun) []components.CodeEditorRun {
	tag := vivaui.RGB24(0x56a8f5)
	comment := vivaui.RGB24(0x7a7e85)
	if appearance == AppearanceLight {
		tag = vivaui.RGB24(0x00627a)
		comment = vivaui.RGB24(0x6c707e)
	}
	for index := 0; index < len(line); {
		switch {
		case strings.HasPrefix(line[index:], "<!--"):
			end := strings.Index(line[index+4:], "-->")
			if end < 0 {
				return append(dst, components.CodeEditorRun{Start: index, End: len(line), Color: comment})
			}
			end += index + 7
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: comment})
			index = end
		case line[index] == '<':
			end := strings.IndexByte(line[index:], '>')
			if end < 0 {
				return append(dst, components.CodeEditorRun{Start: index, End: len(line), Color: tag})
			}
			end += index + 1
			dst = append(dst, components.CodeEditorRun{Start: index, End: end, Color: tag})
			index = end
		default:
			index++
		}
	}
	return dst
}

func resourceColor(filePath string, theme vivaui.Theme) vivaui.Color {
	switch strings.ToLower(filepath.Ext(filePath)) {
	case ".xml", ".xsd", ".html", ".htm":
		return vivaui.RGB24(0xd5b778)
	case ".json", ".yaml", ".yml", ".toml":
		return vivaui.RGB24(0x56a8f5)
	default:
		return theme.MutedForeground
	}
}
