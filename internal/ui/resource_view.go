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
	tag := vivaui.RGB24(0xd5b778)
	comment := vivaui.RGB24(0x7a7e85)
	if appearance == AppearanceLight {
		tag = vivaui.RGB24(0x0033b3)
		comment = vivaui.RGB24(0x8c8c8c)
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
